import pytest
import time
from hermes_plugin.context import step, attachment, AttachmentTypeEnum


class TestMockBusiness:
    def setup_class(self):
        with step("Open chrome browser"):
            time.sleep(1)
        with step("Open https://www.taobao.com"):
            time.sleep(1)
        with step("Login with username and password"):
            time.sleep(1)

    def teardown_class(self):
        with step("Logout from https://www.taobao.com"):
            time.sleep(1)
        with step("Close chrome browser"):
            time.sleep(1)

    def setup_method(self):
        with step("Search for 'iphone 15' on taobao"):
            time.sleep(1)

    def teardown_method(self):
        with step("Clear search history"):
            time.sleep(1)

    def test_search_iphone_15(self):
        with step("Search for 'iphone 15' on taobao"):
            time.sleep(1)
        with step("Click on the first result"):
            time.sleep(1)
        with step("Add to cart"):
            time.sleep(1)
        with step("Checkout"):
            time.sleep(1)
        with step("Pay for the order"):
            time.sleep(1)

    def test_search_iphone_15_and_add_to_cart(self):
        with step("Search for 'iphone 15' on taobao"):
            time.sleep(1)
            with step("Click on the first result"):
                time.sleep(1)
                with step("Add to cart"):
                    time.sleep(1)
                    with step("Checkout"):
                        time.sleep(1)
        with step("Pay for the order"):
            time.sleep(1)
        with step("Complete the purchase"):
            time.sleep(1)

    def test_search_iphone_15_and_add_to_cart_and_checkout(self):
        with step("Search for 'iphone 15' on taobao"):
            time.sleep(1)
            with step("Click on the first result"):
                time.sleep(1)
                with step("Add to cart"):
                    time.sleep(1)
                    with step("Checkout"):
                        time.sleep(1)
        with step("Pay for the order"):
            time.sleep(1)
        with step("Complete the purchase"):
            time.sleep(1)

    def test_search_iphone_15_and_add_to_cart_and_checkout_and_pay(self):
        with step("Search for 'iphone 15' on taobao"):
            time.sleep(1)
            with step("Click on the first result"):
                time.sleep(1)
                with step("Add to cart"):
                    time.sleep(1)
                    with step("Checkout"):
                        time.sleep(1)
        with step("Pay for the order"):
            time.sleep(1)
        with step("Complete the purchase"):
            time.sleep(1)

    def test_search_iphone_15_and_add_to_cart_and_checkout_and_pay_and_complete_purchase(
        self,
    ):
        with step("Search for 'iphone 15' on taobao"):
            time.sleep(1)
            with step("Click on the first result"):
                time.sleep(1)
                with step("Add to cart"):
                    time.sleep(1)
                    with step("Checkout"):
                        time.sleep(1)
        with step("Pay for the order"):
            time.sleep(1)
        with step("Complete the purchase"):
            time.sleep(1)

    def test_attachment(self):
        with step("Search for 'iphone 15' on taobao"):
            time.sleep(1)
            with step("Click on the first result"):
                time.sleep(1)
                with step("Add to cart"):
                    time.sleep(1)
                    with step("Checkout"):
                        time.sleep(1)
        with step("Pay for the order"):
            time.sleep(1)
        with step("Complete the purchase"):
            time.sleep(1)
        attachment(
            "Search for 'iphone 15' on taobao", "iphone 15", AttachmentTypeEnum.LOG
        )
        attachment("Click on the first result", "iphone 15", AttachmentTypeEnum.LOG)
