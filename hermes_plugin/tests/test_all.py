import pytest
from hermes_plugin.context import step, attachment, AttachmentTypeEnum

pytestmark = pytest.mark.skip(reason="manual Allure demo, not an automated assertion suite")


class TestAllProcess:
    def setup_class(self):
        with step("Setup class for TestAllProcess"):
            pass

    def teardown_class(self):
        with step("Teardown class for TestAllProcess"):
            pass

    def setup_method(self):
        with step("Setup method for TestAllProcess"):
            pass

    def teardown_method(self):
        with step("Teardown method for TestAllProcess"):
            pass

    def test_step_with_attachment(self):
        with step("Test step with attachment"):
            attachment(
                "Sample Attachment", "This is a sample attachment content.", AttachmentTypeEnum.LOG
            )
            assert True

    def test_step_with_nested_steps(self):
        with step("Test step with nested steps"):
            with step("Nested Step 1"):
                assert True
            with step("Nested Step 2"):
                assert True

    def test_step_with_exception(self):
        with step("Test step with exception"):
            try:
                raise ValueError("This is a test exception.")
            except ValueError as e:
                attachment("Exception Attachment", str(e), AttachmentTypeEnum.LOG)
                assert False, "Intentional failure to test exception handling."
