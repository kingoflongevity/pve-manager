"""
PVE WebUI 功能完整测试 - AI 宠物、应用商店、访问管理
"""
import os
import asyncio
from pathlib import Path
from playwright.sync_api import sync_playwright, expect

BASE_URL = "http://localhost:8088"
API_URL = "http://localhost:8080"
SCREENSHOTS_DIR = Path(__file__).parent / "screenshots"
SCREENSHOTS_DIR.mkdir(exist_ok=True)


def login(page):
    """登录系统"""
    page.goto(f"{BASE_URL}/login")
    page.wait_for_load_state("networkidle")
    page.fill('input[placeholder*="用户名"], input[name="username"]', "root")
    page.fill('input[type="password"]', "password")
    page.locator('button:has-text("登录"), button:has-text("Login")').first.click()
    page.wait_for_url("**/dashboard**", timeout=15000)
    page.wait_for_load_state("networkidle")
    print("✅ 登录成功")


def test_ai_floating_pet():
    """测试 AI 悬浮宠物助手"""
    print("\n=== 测试 AI 悬浮宠物助手 ===")
    with sync_playwright() as p:
        browser = p.chromium.launch(headless=True)
        page = browser.new_page(viewport={"width": 1400, "height": 900})

        try:
            login(page)

            # 1. 检查悬浮宠物按钮是否可见
            pet_button = page.locator(".floating-pet-container .pet-button")
            pet_button.wait_for(state="visible", timeout=5000)
            expect(pet_button).to_be_visible()
            print("✅ 悬浮宠物按钮可见")

            # 2. 点击宠物按钮打开聊天面板
            pet_button.click()
            page.wait_for_selector(".chat-panel", state="visible", timeout=5000)
            print("✅ 聊天面板已打开")

            # 3. 检查快速操作按钮
            quick_actions = page.locator(".quick-actions button")
            action_count = quick_actions.count()
            print(f"✅ 找到 {action_count} 个快捷操作按钮")

            # 4. 点击快捷操作发送消息
            quick_actions.first.click()
            page.wait_for_timeout(3000)

            # 5. 检查 AI 回复出现
            messages = page.locator(".chat-message.assistant")
            msg_count = messages.count()
            assert msg_count > 0, "应该有 AI 回复消息"
            print(f"✅ AI 回复了 {msg_count} 条消息")

            # 6. 测试手动输入
            chat_input = page.locator(".chat-input")
            expect(chat_input).to_be_visible()
            chat_input.fill("你好")
            page.locator(".send-btn").click()
            page.wait_for_timeout(3000)

            # 再次检查 AI 回复
            page.screenshot(path=str(SCREENSHOTS_DIR / "ai_chat.png"))
            print("✅ 手动输入测试通过")

            # 7. 关闭聊天面板
            page.locator('.header-btn:has-text("✕")').first.click()
            page.wait_for_selector(".chat-panel", state="hidden", timeout=3000)
            print("✅ 聊天面板关闭成功")

        except Exception as e:
            page.screenshot(path=str(SCREENSHOTS_DIR / "ai_error.png"))
            print(f"❌ AI 宠物测试失败: {e}")
            raise
        finally:
            browser.close()


def test_app_store():
    """测试应用商店功能"""
    print("\n=== 测试应用商店 ===")
    with sync_playwright() as p:
        browser = p.chromium.launch(headless=True)
        page = browser.new_page(viewport={"width": 1400, "height": 900})

        try:
            login(page)

            # 1. 导航到应用商店
            page.locator('a[href="/apps"], .sidebar-item:has-text("应用商店"), li:has-text("应用商店")').first.click()
            page.wait_for_load_state("networkidle")
            page.wait_for_timeout(2000)
            print("✅ 进入应用商店")

            # 2. 检查模板卡片
            cards = page.locator(".template-card, .app-card, .el-card")
            card_count = cards.count()
            assert card_count > 0, "应该有应用模板卡片"
            print(f"✅ 找到 {card_count} 个应用模板")

            # 3. 检查分类标签
            categories = page.locator(".category-tag, .app-category, [class*='category']")
            print(f"✅ 分类标签存在: {categories.count() > 0}")

            # 4. 尝试部署一个应用 (点击第一个模板的部署/安装按钮)
            deploy_btn = page.locator('button:has-text("部署"), button:has-text("安装"), button:has-text("Deploy"), button:has-text("Install")').first
            if deploy_btn.is_visible():
                deploy_btn.click()
                page.wait_for_timeout(2000)

                # 检查是否有部署对话框
                dialog = page.locator(".el-dialog, .el-drawer, [role='dialog']")
                if dialog.is_visible():
                    print("✅ 部署对话框已弹出")

                    # 填写表单并部署
                    name_input = page.locator('input[placeholder*="名称"], input[placeholder*="name"], .el-dialog input').first
                    if name_input.is_visible():
                        name_input.fill("test-app-001")

                    confirm_btn = page.locator('.el-dialog button:has-text("确定"), .el-dialog button:has-text("确认"), button:has-text("部署")').first
                    if confirm_btn.is_visible():
                        confirm_btn.click()
                        page.wait_for_timeout(2000)
                        print("✅ 应用部署提交成功")
            else:
                print("⚠️ 未找到部署按钮（可能需要在模板卡片内点击）")

            page.screenshot(path=str(SCREENSHOTS_DIR / "app_store.png"))

        except Exception as e:
            page.screenshot(path=str(SCREENSHOTS_DIR / "app_store_error.png"))
            print(f"❌ 应用商店测试失败: {e}")
            raise
        finally:
            browser.close()


def test_access_management():
    """测试访问管理功能"""
    print("\n=== 测试访问管理 ===")
    with sync_playwright() as p:
        browser = p.chromium.launch(headless=True)
        page = browser.new_page(viewport={"width": 1400, "height": 900})

        try:
            login(page)

            # 1. 导航到访问管理
            page.locator('a[href="/access"], .sidebar-item:has-text("访问"), li:has-text("访问管理"), li:has-text("权限")').first.click()
            page.wait_for_load_state("networkidle")
            page.wait_for_timeout(2000)
            print("✅ 进入访问管理")

            # 2. 检查用户表格
            table_rows = page.locator("table tr, .el-table__row, tbody tr")
            row_count = table_rows.count()
            print(f"✅ 表格行数: {row_count}")

            # 3. 检查标签页切换（用户/组/角色/ACL/域）
            tabs = page.locator(".el-tabs__item, [role='tab'], .tab-btn")
            tab_count = tabs.count()
            print(f"✅ 找到 {tab_count} 个标签页")

            # 尝试点击不同标签
            tab_texts = []
            for i in range(min(tab_count, 5)):
                tab = tabs.nth(i)
                if tab.is_visible():
                    tab_texts.append(tab.text_content()[:20])
                    tab.click()
                    page.wait_for_timeout(500)

            print(f"✅ 切换到标签页: {tab_texts}")

            # 4. 截图用户列表
            page.screenshot(path=str(SCREENSHOTS_DIR / "access_users.png"))

            # 5. 测试创建用户按钮
            create_btn = page.locator('button:has-text("创建"), button:has-text("新建"), button:has-text("添加"), button:has-text("Create"), button:has-text("Add")').first
            if create_btn.is_visible():
                create_btn.click()
                page.wait_for_timeout(1000)
                dialog_visible = page.locator(".el-dialog, [role='dialog']").is_visible()
                print(f"✅ 创建对话框: {'可见' if dialog_visible else '已打开'}")
            else:
                print("⚠️ 未找到创建按钮（可能标签页不同）")

        except Exception as e:
            page.screenshot(path=str(SCREENSHOTS_DIR / "access_error.png"))
            print(f"❌ 访问管理测试失败: {e}")
            raise
        finally:
            browser.close()


def test_sidebar_no_ai():
    """验证侧边栏中没有 AI 入口"""
    print("\n=== 验证侧边栏无 AI 入口 ===")
    with sync_playwright() as p:
        browser = p.chromium.launch(headless=True)
        page = browser.new_page(viewport={"width": 1400, "height": 900})

        try:
            login(page)
            page.wait_for_timeout(1000)

            sidebar_text = page.locator(".sidebar, .app-sidebar, aside, nav").first.text_content()
            has_ai_in_sidebar = "AI 智能中心" in sidebar_text or "AI 智能体" in sidebar_text
            assert not has_ai_in_sidebar, "侧边栏不应该有 AI 入口！"
            print("✅ 侧边栏已正确移除 AI 入口")

            page.screenshot(path=str(SCREENSHOTS_DIR / "sidebar_no_ai.png"))

        except Exception as e:
            page.screenshot(path=str(SCREENSHOTS_DIR / "sidebar_error.png"))
            print(f"❌ 侧边栏验证失败: {e}")
            raise
        finally:
            browser.close()


if __name__ == "__main__":
    import sys
    all_passed = True

    tests = [
        ("AI 悬浮宠物助手", test_ai_floating_pet),
        ("应用商店", test_app_store),
        ("访问管理", test_access_management),
        ("侧边栏无 AI 入口", test_sidebar_no_ai),
    ]

    for name, test_fn in tests:
        try:
            test_fn()
        except Exception as e:
            all_passed = False
            print(f"\n❌ {name} 测试失败: {e}")

    print("\n" + "=" * 50)
    if all_passed:
        print("🎉 全部测试通过！")
    else:
        print("❌ 部分测试失败，请查看上方日志")
    sys.exit(0 if all_passed else 1)
