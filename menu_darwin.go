//go:build darwin

// macOS 原生菜单栏（v1.7.45）
//
// webview_go 创建 NSWindow 时不会装 NSMenu，导致 Cmd+C/V/X/A/Z 这种标准 shortcut
// 在 OS 层面没有 menu item 路由——JS keydown 能不能拿到取决于 WKWebView 版本，
// 不可靠。
//
// 装一个标准的 NSMenu（含 App / Edit 两个子菜单），让 OS 把
// Cmd+C/V/X/A/Z/Q 路由到 first responder（聚焦的文本框）。
// 这是 macOS 标准做法，100% 可靠。

package main

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework Cocoa
#import <Cocoa/Cocoa.h>

static void addItem(NSMenu *menu, NSString *title, SEL action, NSString *keyEq, NSEventModifierFlags mods) {
    NSMenuItem *item = [[NSMenuItem alloc] initWithTitle:title action:action keyEquivalent:keyEq];
    if (mods != 0) {
        [item setKeyEquivalentModifierMask:mods];
    }
    [menu addItem:item];
}

void installAppMenu(void) {
    @autoreleasepool {
        NSApplication *app = [NSApplication sharedApplication];

        NSMenu *mainMenu = [[NSMenu alloc] init];

        // ===== App 菜单（第一个，标题被 macOS 替换为 App 名）=====
        NSMenuItem *appMenuItem = [[NSMenuItem alloc] init];
        [mainMenu addItem:appMenuItem];
        NSMenu *appMenu = [[NSMenu alloc] init];

        NSString *appName = @"Team Standards";

        addItem(appMenu, [NSString stringWithFormat:@"About %@", appName], NULL, @"", 0);
        [appMenu addItem:[NSMenuItem separatorItem]];
        addItem(appMenu, [NSString stringWithFormat:@"Hide %@", appName], @selector(hide:), @"h", 0);

        NSMenuItem *hideOthers = [[NSMenuItem alloc] initWithTitle:@"Hide Others" action:@selector(hideOtherApplications:) keyEquivalent:@"h"];
        [hideOthers setKeyEquivalentModifierMask:NSEventModifierFlagCommand | NSEventModifierFlagOption];
        [appMenu addItem:hideOthers];

        addItem(appMenu, @"Show All", @selector(unhideAllApplications:), @"", 0);
        [appMenu addItem:[NSMenuItem separatorItem]];
        addItem(appMenu, [NSString stringWithFormat:@"Quit %@", appName], @selector(terminate:), @"q", 0);
        [appMenuItem setSubmenu:appMenu];

        // ===== Edit 菜单（关键：Cmd+C/V/X/A/Z 等）=====
        NSMenuItem *editMenuItem = [[NSMenuItem alloc] init];
        [mainMenu addItem:editMenuItem];
        NSMenu *editMenu = [[NSMenu alloc] initWithTitle:@"Edit"];

        addItem(editMenu, @"Undo", @selector(undo:), @"z", 0);

        NSMenuItem *redoItem = [[NSMenuItem alloc] initWithTitle:@"Redo" action:@selector(redo:) keyEquivalent:@"z"];
        [redoItem setKeyEquivalentModifierMask:NSEventModifierFlagCommand | NSEventModifierFlagShift];
        [editMenu addItem:redoItem];

        [editMenu addItem:[NSMenuItem separatorItem]];

        addItem(editMenu, @"Cut", @selector(cut:), @"x", 0);
        addItem(editMenu, @"Copy", @selector(copy:), @"c", 0);
        addItem(editMenu, @"Paste", @selector(paste:), @"v", 0);
        addItem(editMenu, @"Delete", @selector(delete:), @"", 0);
        addItem(editMenu, @"Select All", @selector(selectAll:), @"a", 0);

        [editMenuItem setSubmenu:editMenu];

        // ===== Window 菜单（Cmd+M / Cmd+W）=====
        NSMenuItem *winMenuItem = [[NSMenuItem alloc] init];
        [mainMenu addItem:winMenuItem];
        NSMenu *winMenu = [[NSMenu alloc] initWithTitle:@"Window"];
        addItem(winMenu, @"Minimize", @selector(performMiniaturize:), @"m", 0);
        addItem(winMenu, @"Close Window", @selector(performClose:), @"w", 0);
        [winMenuItem setSubmenu:winMenu];
        [app setWindowsMenu:winMenu];

        [app setMainMenu:mainMenu];
    }
}
*/
import "C"

// installNativeAppMenu 在主线程上安装 macOS 标准菜单栏。
// 必须在 webview.New 之前调用，确保 NSApplication 已初始化但 NSMenu 还没起来。
//
// 副作用：
//   - 用户能用 Cmd+C/V/X/A/Z 标准编辑快捷键（OS 路由到 first responder）
//   - 用户能用 Cmd+Q 退出（与 JS handler 兼容，两条路都行）
//   - 用户能用 Cmd+M 最小化、Cmd+W 关窗口
//   - 顶部菜单栏出现「Team Standards · Edit · Window」3 组
func installNativeAppMenu() {
	C.installAppMenu()
}
