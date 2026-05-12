"""
PVE WebUI 浏览器自动化测试脚本
使用 Playwright 进行全面的浏览器测试和截图
"""
import os
import sys
import subprocess
from pathlib import Path


def run_playwright_tests():
    project_root = Path(__file__).parent.parent
    frontend_dir = project_root / "frontend"

    os.chdir(str(frontend_dir))
    print(f"工作目录: {os.getcwd()}")

    result = subprocess.run(
        "npx playwright test --reporter=list,html",
        capture_output=True,
        text=True,
        timeout=300,
        shell=True,
    )

    print("=== STDOUT ===")
    print(result.stdout)
    print("=== STDERR ===")
    print(result.stderr)
    print(f"=== 退出码: {result.returncode} ===")

    return result.returncode


if __name__ == "__main__":
    sys.exit(run_playwright_tests())
