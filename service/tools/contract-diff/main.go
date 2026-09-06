// Command contract-diff 比对 Go 路由表与 Python router.py 的接口一致性。
//
// 用途（T-10.1）：确保 Go 侧 `internal/adapter/http/router.go` 注册的路径/方法
// 与 Python 侧 `service/src/app/**/router.py`（含子 router 前缀）保持一致。
// 不一致时退出码非 0，供 CI 门禁调用。
//
// 说明：
//   - Python 侧 auth 拆成多个子 router（users/roles/permissions），各自带 prefix；
//     本工具解析 `@xxx_router.METHOD("path")` 装饰器并叠加 prefix。
//   - Go 侧路由统一由 `internal/adapter/http/router.go` 的 `registerXxxRoutes`
//     注册，本工具按方法名 + 路径字面量提取。
//   - 前缀 `API_V1_STR` = `/tap/api/v1` 由调用方传入（-api-prefix），
//     Go 侧同样的前缀（`cfg.HTTP.APIPrefix`）。比对时都剥离前缀，只看相对路径。
//
// 用法：
//
//	go run ./tools/contract-diff -python ../service/src/app -go ../go-service/internal/adapter/http
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// Endpoint 是一条接口记录（方法 + 相对路径）。
type Endpoint struct {
	Method string
	Path   string
}

func (e Endpoint) String() string { return e.Method + " " + e.Path }

// ---------------------------------------------------------------------------
// Python 解析
// ---------------------------------------------------------------------------

var (
	pyRouterDeclRe = regexp.MustCompile(`(\w+)\s*=\s*APIRouter\([^)]*prefix\s*=\s*"([^"]*)"`)
	pyDecoratorRe  = regexp.MustCompile(`@(\w+)\.(get|post|put|patch|delete)\(\s*\n?\s*["']([^"']*)["']`)
	pyPathParamRe  = regexp.MustCompile(`\{([^}]+)\}`)
	pyIncludeRe    = regexp.MustCompile(`include_router\((\w+)\s*,\s*prefix\s*=\s*settings\.(\w+)`)
)

// pythonRouters 解析 Python 目录，返回活跃 router 的 prefix -> 接口集。
func pythonRouters(pythonDir string, apiPrefix string) (map[string][]Endpoint, error) {
	// 第一步：收集所有 router 声明（router 变量名 -> prefix）
	decls := map[string]string{}       // routerName -> prefix
	activeRouters := map[string]bool{} // 被 main include 的 router 名

	mainFile := filepath.Join(pythonDir, "app", "main.py")
	if mainContent, err := os.ReadFile(mainFile); err == nil {
		for _, m := range pyIncludeRe.FindAllStringSubmatch(string(mainContent), -1) {
			if len(m) == 3 {
				// m[1] = router 名, m[2] = settings 变量名（API_V1_STR）
				activeRouters[m[1]] = true
			}
		}
	}

	// 递归扫描所有 .py 文件
	err := filepath.Walk(pythonDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() || !strings.HasSuffix(path, ".py") {
			return nil
		}
		content, err := os.ReadFile(path)
		if err != nil {
			return nil
		}
		text := string(content)

		// 声明 router 前缀
		for _, m := range pyRouterDeclRe.FindAllStringSubmatch(text, -1) {
			if len(m) == 3 {
				decls[m[1]] = m[2]
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	// 第二步：收集所有装饰器定义的接口，叠加 router prefix
	result := map[string][]Endpoint{}
	seen := map[string]Endpoint{} // 去重 key: routerName|method|path

	_ = filepath.Walk(pythonDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || !strings.HasSuffix(path, ".py") {
			return nil
		}
		content, err := os.ReadFile(path)
		if err != nil {
			return nil
		}
		text := string(content)

		for _, m := range pyDecoratorRe.FindAllStringSubmatch(text, -1) {
			if len(m) != 4 {
				continue
			}
			routerName := m[1]
			method := strings.ToUpper(m[2])
			path := m[3]
			// path 参数可能是 f-string 或含 /{}；这里取第一段字符串字面量
			path = firstStringLiteral(path)

			prefix := decls[routerName]
			fullPath := normalizePath(prefix + path)

			key := routerName + "|" + method + "|" + fullPath
			if _, dup := seen[key]; dup {
				continue
			}
			seen[key] = Endpoint{Method: method, Path: fullPath}

			// 只收集被 include 的 router（活跃），或前缀非空的子 router 也在文档契约里
			// 这里统一收集，最终与契约文档交叉核对由调用方判断
			result[routerName] = append(result[routerName], Endpoint{Method: method, Path: fullPath})
		}
		return nil
	})

	return result, nil
}

// toNativePath 把 Git Bash 风格的 Unix 路径（/d/workspace/...）转成
// Windows 原生路径（D:\workspace\...），filepath.Walk 需要后者。
// 已是 Windows 风格（含盘符冒号）或 Linux 路径时原样返回。
func toNativePath(p string) string {
	if len(p) >= 3 && p[0] == '/' && p[2] == '/' {
		// /d/foo → D:\foo
		drive := p[1]
		if (drive >= 'a' && drive <= 'z') || (drive >= 'A' && drive <= 'Z') {
			return strings.ToUpper(string(drive)) + ":" + filepath.FromSlash(p[2:])
		}
	}
	return p
}

// firstStringLiteral 从一段 Python 表达式里取出第一个字符串字面量。
// 处理 "..." 与 '...'，跳过 f 前缀。
func firstStringLiteral(s string) string {
	// 去掉 f 前缀（如 f"/a/{b}"）
	s = strings.TrimSpace(s)
	if len(s) > 0 && (s[0] == 'f' || s[0] == 'F' || s[0] == 'r' || s[0] == 'b') &&
		len(s) > 1 && (s[1] == '"' || s[1] == '\'') {
		s = s[1:]
	}
	s = strings.TrimSpace(s)
	if len(s) >= 2 {
		q := s[0]
		if q == '"' || q == '\'' {
			// 找到闭合引号
			for i := 1; i < len(s); i++ {
				if s[i] == q && s[i-1] != '\\' {
					return s[1:i]
				}
			}
		}
	}
	return s
}

// normalizePath 统一完整端点路径：把 {param} 归一为 :param、去掉结尾斜杠。
func normalizePath(p string) string {
	p = strings.TrimSpace(p)
	p = pyPathParamRe.ReplaceAllString(p, ":$1")
	p = strings.TrimRight(p, "/")
	if p == "" {
		p = "/"
	}
	return p
}

// normalizeGroupPath 归一化 group 前缀：空串保持空（表示无子路径的顶层 group）。
func normalizeGroupPath(p string) string {
	p = strings.TrimSpace(p)
	p = pyPathParamRe.ReplaceAllString(p, ":$1")
	return strings.Trim(p, "/")
}

// ---------------------------------------------------------------------------
// Go 解析
// ---------------------------------------------------------------------------

var (
	goRouterRe = regexp.MustCompile(`\.(GET|POST|PUT|PATCH|DELETE)\("([^"]*)"`)
	goGroupRe  = regexp.MustCompile(`(\w+)\s*:?=\s*(\w+)\.Group\("([^"]*)"`)
)

// goRouters 解析 Go router.go 的注册。
//
// Go 路由用 `X := api.Group("/automation")` 嵌套注册，方法调用
// `query.GET("/executions", ...)` 的路径要叠加所在 Group 的 prefix。
// 这里按行扫描：遇到 Group 声明记录其前缀，遇到方法注册叠加最近的 group 前缀。
func goRouters(goDir string) ([]Endpoint, error) {
	routerFile := filepath.Join(goDir, "router.go")
	content, err := os.ReadFile(routerFile)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", routerFile, err)
	}
	lines := strings.Split(string(content), "\n")

	groupPrefix := map[string]string{} // group 变量名 -> 前缀
	var out []Endpoint
	seen := map[string]bool{}

	for _, line := range lines {
		// Group 声明：X := parent.Group("/path")
		if m := goGroupRe.FindStringSubmatch(line); len(m) == 4 {
			name, parent, path := m[1], m[2], m[3]
			base := ""
			if parent != "api" && parent != "engine" && parent != "authed" {
				base = groupPrefix[parent]
			}
			groupPrefix[name] = normalizeGroupPath(base + path)
			continue
		}

		// 方法注册
		if m := goRouterRe.FindStringSubmatch(line); len(m) == 3 {
			method := strings.ToUpper(m[1])
			path := normalizePath(m[2])
			if path == "/healthz" || path == "/readyz" {
				continue
			}
			// 找该行最近的 group：从行内容里找当前所属 group 变量。
			// 方法注册行形如 `system.GET(...)` 或 `query.GET(...)`，其接收者
			// 即 group 变量名。取 `.METHOD(` 前的标识符。
			prefix := ""
			if dot := regexp.MustCompile(`([A-Za-z_]\w*)\.` + m[1] + `\(`).FindStringSubmatch(line); len(dot) == 2 {
				prefix = groupPrefix[dot[1]]
			}
			full := prefix + path
			if !strings.HasPrefix(full, "/") {
				full = "/" + full
			}
			full = normalizePath(full)
			e := Endpoint{Method: method, Path: full}
			k := e.Method + " " + e.Path
			if !seen[k] {
				seen[k] = true
				out = append(out, e)
			}
		}
	}
	return out, nil
}

// ---------------------------------------------------------------------------
// 主逻辑
// ---------------------------------------------------------------------------

func main() {
	pythonDir := flag.String("python", "../../service/src/app", "Python app 目录")
	goDir := flag.String("go", ".", "Go adapter/http 目录")
	apiPrefix := flag.String("api-prefix", "/tap/api/v1", "API 统一前缀（剥离比较）")
	flag.Parse()

	pythonRoutes, err := pythonRouters(toNativePath(*pythonDir), *apiPrefix)
	if err != nil {
		log.Fatalf("parse python routers: %v", err)
	}
	goRoutes, err := goRouters(toNativePath(*goDir))
	if err != nil {
		log.Fatalf("parse go router: %v", err)
	}

	// 汇总 Python 全部接口（所有 router）
	var pyEndpoints []Endpoint
	for _, eps := range pythonRoutes {
		pyEndpoints = append(pyEndpoints, eps...)
	}
	pyEndpoints = dedupEndpoints(pyEndpoints)

	// 剥离前缀后比较
	pySet := map[string]Endpoint{}
	for _, e := range pyEndpoints {
		rel := strings.TrimPrefix(e.Path, *apiPrefix)
		rel = normalizePath(rel)
		pySet[rel] = Endpoint{Method: e.Method, Path: rel}
	}
	goSet := map[string]Endpoint{}
	for _, e := range goRoutes {
		rel := normalizePath(e.Path)
		goSet[rel] = Endpoint{Method: e.Method, Path: rel}
	}

	var missing, extra []string
	for k, pe := range pySet {
		if _, ok := goSet[k]; !ok {
			missing = append(missing, "Python 有 Go 缺: "+pe.String())
		}
	}
	for k, ge := range goSet {
		if _, ok := pySet[k]; !ok {
			// 设计文档确认的 Go 侧新增接口不计为差异（见 docs/go-service/03-api-contract.md）
			if allowedGoExtra(ge) {
				continue
			}
			extra = append(extra, "Go 有 Python 缺: "+ge.String())
		}
	}

	sort.Strings(missing)
	sort.Strings(extra)

	fmt.Printf("Python 接口数: %d, Go 接口数: %d\n", len(pySet), len(goSet))
	if len(missing) > 0 {
		fmt.Println("--- 差异：Python 有而 Go 缺 ---")
		for _, m := range missing {
			fmt.Println("  " + m)
		}
	}
	if len(extra) > 0 {
		fmt.Println("--- 差异：Go 有而 Python 缺（非白名单）---")
		for _, m := range extra {
			fmt.Println("  " + m)
		}
	}

	if len(missing) > 0 || len(extra) > 0 {
		fmt.Println("contract-diff: MISMATCH")
		os.Exit(1)
	}
	fmt.Println("contract-diff: OK")
}

// allowedGoExtra 返回设计文档确认的 Go 侧新增接口（Python 无对应实现，
// 但契约明确要求 Go 实现）。这些不算契约不一致。
func allowedGoExtra(e Endpoint) bool {
	allowed := map[string]bool{
		// Go 新增 /login/refresh（Python 有 Redis refresh store 但未暴露端点）
		"POST /login/refresh": true,
		"GET /automation/dashboard/overview":                       true,
		"GET /automation/dashboard/executions/:execution_id/cases": true,
		"POST /system/dicts":       true,
		"PUT /system/dicts/:id":    true,
		"DELETE /system/dicts/:id": true,
		"POST /automation/executions/:build_uid/heartbeat": true,
		"POST /automation/items/:case_uid/steps":           true,
		"GET /automation/executions/:build_uid/events":     true,
		"GET /automation/executions/:build_uid/live":       true,
		"GET /automation/executions/:build_uid/items":      true,
		"GET /automation/items/:case_uid":                  true,
	}
	return allowed[e.String()]
}

func dedupEndpoints(eps []Endpoint) []Endpoint {
	seen := map[string]bool{}
	var out []Endpoint
	for _, e := range eps {
		k := e.Method + " " + e.Path
		if !seen[k] {
			seen[k] = true
			out = append(out, e)
		}
	}
	return out
}
