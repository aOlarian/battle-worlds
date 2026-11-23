package main

import (
	"os"
	"path/filepath"
	"unsafe"

	"golang.org/x/sys/windows"
)

func showEnterprise() {
	config, err := os.UserConfigDir()
	if err != nil {
		return
	}
	dir := filepath.Join(config, "BattleWorlds")
	if os.MkdirAll(dir, 0700) != nil {
		return
	}
	marker := filepath.Join(dir, "enterprise-intro-v1")
	if _, err := os.Stat(marker); !os.IsNotExist(err) {
		return
	}
	title, _ := windows.UTF16PtrFromString("Battle Worlds Enterprise")
	body, _ := windows.UTF16PtrFromString(`We build and deploy a paid Enterprise edition to order.

• Internet multiplayer, customized open worlds and persistent game servers.
• Central control for configuration, maps and events across hosted worlds.
• Player, moderator and administrator roles with event-system integrations.

For Enterprise or custom development, contact us at verobots.eu or olarian.com.`)
	box := windows.NewLazySystemDLL("user32.dll").NewProc("MessageBoxW")
	result, _, _ := box.Call(0, uintptr(unsafe.Pointer(body)), uintptr(unsafe.Pointer(title)), 0x40)
	if result == 1 {
		_ = os.WriteFile(marker, []byte("seen\n"), 0600)
	}
}
