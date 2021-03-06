package lwip

/*
#cgo CFLAGS: -I./src/include
#include "lwip/init.h"
#include "lwip/tcp.h"
#include "lwip/udp.h"
*/
import "C"
import (
	"log"
	"sync"
	"unsafe"
)

var listenTCPPCB *C.struct_tcp_pcb
var listenUDPPCB *C.struct_udp_pcb
var lwipMutex = &sync.Mutex{}
var udpMutex = &sync.Mutex{}

func Setup() {
	C.lwip_init()

	tcpPCB := C.tcp_new()
	if tcpPCB == nil {
		panic("tcp_new return nil")
	}

	err := C.tcp_bind(tcpPCB, &C.ip_addr_any, 0)
	switch err {
	case C.ERR_OK:
		break
	case C.ERR_VAL:
		log.Fatal("invalid PCB state")
	case C.ERR_USE:
		log.Fatal("port in use")
	default:
		C.memp_free(C.MEMP_TCP_PCB, unsafe.Pointer(tcpPCB))
		log.Fatal("unknown tcp_bind return value")
	}

	tcpPCB = C.tcp_listen_with_backlog(tcpPCB, C.TCP_DEFAULT_LISTEN_BACKLOG)

	// We can't call C function with Go functions as arguments here, it will
	// fail in compile time:
	// cannot use TCPAcceptFn (type func(unsafe.Pointer, *_Ctype_struct_tcp_pcb, _Ctype_schar) _Ctype_schar) as type *[0]byte in argument to func literal
	// I can't find other workarounds.
	// C.tcp_accept(tcpPCB, TCPAcceptFn)
	SetTCPAcceptCallback(tcpPCB)

	listenTCPPCB = tcpPCB

	udpPCB := C.udp_new()
	if udpPCB == nil {
		panic("could not allocate udp pcb")
	}

	err = C.udp_bind(udpPCB, &C.ip_addr_any, 0)
	if err != C.ERR_OK {
		log.Fatal("address already in use")
	}

	SetUDPRecvCallback(udpPCB, nil)
	listenUDPPCB = udpPCB
}
