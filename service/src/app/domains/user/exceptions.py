from fastapi import HTTPException, status


class UserDomainException(HTTPException):
    """用户领域基础异常"""
    def __init__(self, status_code: int, detail: str):
        super().__init__(status_code=status_code, detail=detail)


class InvalidCredentialsException(UserDomainException):
    """账号或密码错误异常"""
    def __init__(self, detail: str = "用户名或密码错误"):
        super().__init__(status_code=status.HTTP_400_BAD_REQUEST, detail=detail)


class UserInactiveException(UserDomainException):
    """账号被禁用异常"""
    def __init__(self, detail: str = "账号已被禁用"):
        super().__init__(status_code=status.HTTP_400_BAD_REQUEST, detail=detail)


class UserNotFoundException(UserDomainException):
    """未找到指定用户异常"""
    def __init__(self, detail: str = "用户不存在"):
        super().__init__(status_code=status.HTTP_404_NOT_FOUND, detail=detail)


class UserAlreadyExistsException(UserDomainException):
    """用户名或邮箱已存在异常"""
    def __init__(self, detail: str = "用户名或邮箱已被注册"):
        super().__init__(status_code=status.HTTP_400_BAD_REQUEST, detail=detail)

