package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime/debug"
	"time"
)

// Jurnal de pornire: %APPDATA%\BattleWorlds\pornire.log, rescris la fiecare pornire.
// Daca jocul se opreste la incarcare, ultima linie arata pasul la care a ramas.
var bootStart = time.Now()

func bootLog(format string, args ...any) {
	path := filepath.Join(filepath.Dir(getSavePath()), "pornire.log")
	flag := os.O_CREATE | os.O_WRONLY | os.O_APPEND
	if format == "start" {
		flag = os.O_CREATE | os.O_WRONLY | os.O_TRUNC
	}
	f, err := os.OpenFile(path, flag, 0644)
	if err != nil {
		return
	}
	defer f.Close()
	fmt.Fprintf(f, "%7.2fs  %s\n", time.Since(bootStart).Seconds(), fmt.Sprintf(format, args...))
}

// crashLog: daca jocul cade (panic), scrie cauza si stiva in pornire.log si crash.log.
func crashLog(r any) {
	msg := fmt.Sprintf("CADERE: %v\n%s", r, debug.Stack())
	bootLog("%s", msg)
	path := filepath.Join(filepath.Dir(getSavePath()), "crash.log")
	_ = os.WriteFile(path, []byte(time.Now().Format("2006-01-02 15:04:05")+"\n"+msg), 0644)
	fmt.Println(msg)
}
