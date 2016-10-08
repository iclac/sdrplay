package sdrplay

import "C"
import (
	"log"
	"unsafe"
)

// StreamCallback è la funzione che viene invocata dall'API SDRplay quando ci sono
// campioni da processare.

//export StreamCallback
func StreamCallback(xi *C.short, xq *C.short, firstSampleNum C.uint, grChanged C.int, rfChanged C.int, fsChanged C.int, numSample C.uint, reset C.uint, cbContext unsafe.Pointer) {
	if grChanged == 1 || fsChanged == 1 || reset == 1 || rx.baseband == nil {
		return
	}

	//fs := int(firstSampleNum)
	//log.Println("fs:", fs)

	is := (*[1 << 30]int16)(unsafe.Pointer(xi))[:numSample:numSample]
	i := make([]int16, len(is))
	copy(i, is)

	qs := (*[1 << 30]int16)(unsafe.Pointer(xi))[:numSample:numSample]
	q := make([]int16, len(qs))
	copy(q, qs)

	rx.baseband.Propagate(i, q)
	//rx.baseband.Propagate(i[fs:], q[fs:])
}

// AGCCallback è la funzione che viene invocata dall'API SDRplay quando ci sono
// variazioni nel guadagno della RSP dovute al AGC.

//export AGCCallback
func AGCCallback(grdB C.uint, lnagrdB C.uint, cbContext unsafe.Pointer) {
	log.Printf("AGC callback [grdB: %d] [lnagrdB: %d]\n", int(grdB), int(lnagrdB))

}
