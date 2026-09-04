from fastapi import HTTPException, status


class AutomationDomainException(HTTPException):
    """自动化用例执行领域基础异常"""

    def __init__(self, status_code: int, detail: str):
        super().__init__(status_code=status_code, detail=detail)


class ExecutionNotFoundException(AutomationDomainException):
    """未找到指定 execution 异常"""

    def __init__(self, detail: str = "execution 不存在"):
        super().__init__(status_code=status.HTTP_404_NOT_FOUND, detail=detail)


class ExecutionItemNotFoundException(AutomationDomainException):
    """未找到指定用例 attempt 异常"""

    def __init__(self, detail: str = "用例执行记录不存在"):
        super().__init__(status_code=status.HTTP_404_NOT_FOUND, detail=detail)


class StepNotFoundException(AutomationDomainException):
    """未找到指定步骤异常"""

    def __init__(self, detail: str = "步骤记录不存在"):
        super().__init__(status_code=status.HTTP_404_NOT_FOUND, detail=detail)
