/*
   sdrplay is a Go package that enables to use the RSP (by SDRplay) in a Go program.
   Copyright (C) 2016 Claudio Carraro carraro.claudio@gmail.com

   See the COPYING file to GPLv2 license details.
*/

package sdrplay

/*

 #cgo CFLAGS: -I/usr/local/include
 #cgo LDFLAGS: -L/usr/local/lib -lmirsdrapi-rsp

 #include "mirsdrapi-rsp.h"
 #include <stdlib.h>

 float api_ver = MIR_SDR_API_VERSION;

 extern void StreamCallback(short *xi, short *xq, unsigned int firstSampleNum, int grChanged, int rfChanged, int fsChanged, unsigned int numSamples, unsigned int reset, void *cbContext);

 extern void AGCCallback(unsigned int grdB, unsigned int lnagrdB, void *cbContext);

 // streamCallback è la funzione che viene invocata dall'API SDRplay quando ci
 // sono campioni da processare.
 static inline void streamCallback(short *xi, short *xq, unsigned int firstSampleNum, int grChanged, int rfChanged, int fsChanged, unsigned int numSamples, unsigned int reset, void *cbContext) {
	StreamCallback(xi, xq, firstSampleNum, grChanged, rfChanged, fsChanged, numSamples, reset, cbContext);
 }

 // agcCallback è la funzione che viene invocata dall'API SDRplay quando ci sono
 // delle variazioni di guadagno nel loop di retroazione del AGC.
 static inline void agcCallback(unsigned int grdB, unsigned int lnagrdB, void *cbContext) {
	AGCCallback(grdB, lnagrdB, cbContext);
 }

 // streamInit è la funzione che invoca l'API mir_sdr_StreamInit.
 mir_sdr_ErrT streamInit(int *gRdB, double fsMHz, double rfMHz, mir_sdr_Bw_MHzT bwType, mir_sdr_If_kHzT ifType, int LNAEnable, int *gRdBsystem, int useGrAltMode, int *samplesPerPacket) {
	return mir_sdr_StreamInit(gRdB, fsMHz, rfMHz, bwType, ifType, LNAEnable, gRdBsystem, useGrAltMode, samplesPerPacket, streamCallback, agcCallback, (void *)NULL);
 }
*/
import "C"
import "log"

// init verifica la versione della libreria, in caso di errore ottenuto dall'API
// o di non corrispondenza di versione viene sollevato un errore fatale.
func init() {
	var vr C.float
	ev := C.mir_sdr_ApiVersion(&vr)
	if ev != 0 {
		log.Fatalf("ApiVersion check Error: %s\n", ev)
	}

	if C.api_ver != vr {
		log.Fatalf("API version mismatch! Version is %f\n", vr)
	}
}

type (
	// radio mantiene lo stato attuale della RSP.
	radio struct {
		// baseband è il connettore dal quale viene propagato il segnale in banda
		// base ricevuto dalla RSP.
		baseband Connector

		// rf è la frequenza attualmente sintonizzata espressa in Hz.
		rf float64

		// band contiene il valore della banda nella quale è sintonizzata la RSP
		// I valori corrispondono con quelli definiti nel enum mir_sdr_BandT.
		band int

		// gr è l'attuale valore di gain reduction
		gr *C.int

		// grsys è il valore del gain reduction del sistema se useGrAltMode == 1
		grsys *C.int

		// spp è il valore di samples per packet
		spp *C.int

		// useGrAltMode ha lo stesso significato dell'API
		useGrAltMode C.int

		// feat contiene le caratteristiche attualmente impostate nella radio.
		feat features
	}

	// enable è un alias di bool introdotto solo per avere una sintassi più
	// comoda per la conversione del relativo valore nel formato compreso dall'API
	enable bool
	// double è un alias di float64 introdotto solo per avere una sintassi più
	// comoda per la conversione del relativo valore nel formato compreso dall'API
	double float64
	// integer è un alias di int introdotto solo per avere una sintassi più
	// comoda per la conversione del relativo valore nel formato compreso dall'API
	integer int

	// features contiene tutti i parametri che si possono configurare nella RSP.
	features struct {
		FS          double
		BW          B
		IF          IFmode
		IQimbalance enable
		DCoffset    enable
		DCmode      OffsetMode
		DCTrakTime  integer
		LOppm       double
		LOmode      LOfrequency
		Decimate    enable
		Factor      Decimation
		LNA         enable
		AGC         AGCmode
		DBFS        integer
		InitialGR   integer
		InitialRF   double
		Debug       enable
	}
)

var (
	// fm102MHz è una configurazione di default che serve nel caso venga invocata
	// la funzione RSP senza alcun parametro di opzione. In particolare la RSP
	// viene impostata per:
	//   * Sintonizzarsi sulla frequenza 102.0 MHz
	//   * Campionare il segnale IF con una FS pari a 2.048 MHz
	//   * Usare una larghezza di banda pari a 1536 kHz
	//   * Usare una IF di 0
	//   * Impostare il modo automatico di gestione della frequenza del up-converter
	fm102MHz = []Option{
		InitialRF(102),
		FS(2.048),
		Bandwidth(BW1536),
		IF(IFzero),
		LOmode(LOauto),
	}
)

var (
	// rsp contiene i valori attuali dei parametri configurati sulla RSP.
	rsp features

	// rx è l'oggetto che rappresenta sempre lo stato attuale della RSP. rx è
	// globale perchè rappresenta un'unica unità RSP.
	rx *radio
)

// newRadio inizializza i puntatori di rx e delle variabili puntatore di radio.
func newRadio() {
	rx = new(radio)

	rx.gr = new(C.int)
	rx.grsys = new(C.int)
	rx.spp = new(C.int)
}

// Tune implementa l'interfaccia Tuner.
func (r *radio) Tune(frequency float64) error {
	if r.baseband == nil {
		return DeactivatedReceiverError
	}

	nb := band(frequency)
	if nb == r.band {
		return toError(C.mir_sdr_SetRf(double(frequency).C(), 1, 0))
	}

	r.band = nb

	var reason C.mir_sdr_ReasonForReinitT = C.mir_sdr_CHANGE_RF_FREQ
	var rfMHz = double(frequency / 1.0e6)

	return toError(C.mir_sdr_Reinit(nil, 0, rfMHz.C(), 0, 0, 0, 0, nil, 0, nil, reason))
}

// Gain implementa l'intarfaccia Amplifier.
func (r *radio) Gain(reduction int) error {
	if r.baseband == nil {
		return DeactivatedReceiverError
	}

	*r.gr = integer(reduction).C()

	return toError(C.mir_sdr_SetGrAltMode(r.gr, C.int(r.feat.LNA.C()), r.grsys, 1, 0))
}

// SetUp implementa l'ultimo metodo dell'interfaccia Receiver così rende radio
// un Receiver.
func (r *radio) SetUp(opts ...Option) error {
	if r.baseband == nil {
		return DeactivatedReceiverError
	}

	configure(opts...)

	if rsp.DCmode != r.feat.DCmode && rsp.DCmode != None {
		C.mir_sdr_SetDcMode(rsp.DCmode.C(), 0)
		C.mir_sdr_SetDcTrackTime(rsp.DCTrakTime.C())
	}

	if rsp.LOppm != r.feat.LOppm && rsp.LOppm != 0.0 {
		C.mir_sdr_SetPpm(rsp.LOppm.C())
	}

	var reason C.mir_sdr_ReasonForReinitT = C.mir_sdr_CHANGE_NONE

	if rsp.InitialGR != r.feat.InitialGR || rsp.LNA != r.feat.LNA {
		reason |= C.mir_sdr_CHANGE_GR
	}

	if rsp.FS != r.feat.FS {
		reason |= C.mir_sdr_CHANGE_FS_FREQ
	}

	if rsp.InitialRF != r.feat.InitialRF {
		reason |= C.mir_sdr_CHANGE_RF_FREQ
	}

	if rsp.BW != r.feat.BW {
		reason |= C.mir_sdr_CHANGE_BW_TYPE
	}

	if rsp.IF != r.feat.IF {
		reason |= C.mir_sdr_CHANGE_IF_TYPE
	}

	if rsp.LOmode != r.feat.LOmode {
		reason |= C.mir_sdr_CHANGE_LO_MODE
	}

	r.feat = rsp

	if reason != C.mir_sdr_CHANGE_NONE {
		*r.gr = r.feat.InitialGR.C()
		*r.grsys = 0
		*r.spp = 0
		r.useGrAltMode = 1

		return toError(C.mir_sdr_Reinit(r.gr, r.feat.FS.C(), r.feat.InitialRF.C(), r.feat.BW.C(), r.feat.IF.C(), r.feat.LOmode.C(), C.int(r.feat.LNA.C()), r.grsys, r.useGrAltMode, r.spp, reason))
	}

	return nil
}

// init inizializza RSP e abilita lo Stream dei campioni in banda base.
func (r *radio) init() error {
	*r.gr = r.feat.InitialGR.C()
	*r.grsys = 0
	*r.spp = 0
	r.useGrAltMode = 1

	// Si abilita o meno il debugging. Non esegue controllo di errore.
	C.mir_sdr_DebugEnable(r.feat.Debug.C())

	// Si abilitano o meno DC offset e IQ imbalance. Non esegue controllo di
	// errore.
	C.mir_sdr_DCoffsetIQimbalanceControl(r.feat.DCoffset.C(), r.feat.IQimbalance.C())

	// Imposta il fattore di decimazione se presente. Non esegue controllo di
	// errore.
	C.mir_sdr_DecimateControl(r.feat.Decimate.C(), r.feat.Factor.C(), 0)

	// Imposta l'AGC: attualmente impone aggiornamento immediato. Non esegue
	// controllo di errore.
	C.mir_sdr_AgcControl(r.feat.AGC.C(), r.feat.DBFS.C(), 0, 0, 0, 0, C.int(r.feat.LNA.C()))

	// Imposta il DC offset mode ed il relativo track time se è stato impostato
	// un DC mode. Non è chiaro dalla documentazione SDRplay se questo valore
	// venga ingnorato nel caso DC offset non sia abilitato, ma penso proprio che
	// sia così.
	if r.feat.DCmode != None {
		C.mir_sdr_SetDcMode(rsp.DCmode.C(), 0)
		C.mir_sdr_SetDcTrackTime(rsp.DCTrakTime.C())
	}

	// Imposta il valore, in parti per milione, del fattore di correzione della
	// frequenza dell'OL della RSP.
	if r.feat.LOppm != 0.0 {
		C.mir_sdr_SetPpm(r.feat.LOppm.C())
	}

	// Imposta il modo di funzionamento del up-converter.
	if r.feat.LOmode != LOundefined {
		C.mir_sdr_SetLoMode(r.feat.LOmode.C())
	}

	dump()

	// LNA è di tipo enable, ma a differenza di tutti gli altri valori che permettono
	// di abilitare una particolare caratteristica che sono di tipo unsigned int,
	// questo è di tipo int. Per questo motivo è necessario il cast a C.int.
	return toError(C.streamInit(r.gr, r.feat.FS.C(), r.feat.InitialRF.C(), r.feat.BW.C(), r.feat.IF.C(), C.int(r.feat.LNA.C()), r.grsys, r.useGrAltMode, r.spp))
}

// uninit ferma lo Stream ed esegue un reset dell'API.
func (r *radio) uninit() error {
	return C.mir_sdr_StreamUninit()
}

// dump mostra su stdout lo stato interno.
func dump() {
	msg := `
--------------------------------------------------------------------------------

	Radio
		%+v


	Features
		%+v

--------------------------------------------------------------------------------
	`

	log.Printf(msg, *rx, rsp)
}

// errDesc mappa i codice di errore delle API SDRplay con le relative descrizioni.
var errDesc = [...]string{
	C.mir_sdr_Success:            "Success",
	C.mir_sdr_Fail:               "Fail",
	C.mir_sdr_InvalidParam:       "Invalid Param",
	C.mir_sdr_OutOfRange:         "Out of Range",
	C.mir_sdr_GainUpdateError:    "Gain Update error",
	C.mir_sdr_RfUpdateError:      "RF Update error",
	C.mir_sdr_FsUpdateError:      "FS Update error",
	C.mir_sdr_HwError:            "HW error",
	C.mir_sdr_AliasingError:      "Aliasing error",
	C.mir_sdr_AlreadyInitialised: "Already Initialised",
	C.mir_sdr_NotInitialised:     "Not Initialised",
}

func (e C.mir_sdr_ErrT) Error() string {
	return errDesc[e]
}

func toError(e C.mir_sdr_ErrT) error {
	if e == C.mir_sdr_Success {
		return nil
	}

	return e
}

// C traduce il valore di e nel formato compreso dall'API SDRplay.
func (e enable) C() C.uint {
	if e {
		return 1
	}

	return 0
}

// C traduce il valore di d nel formato compreso dall'API SDRplay.
func (d double) C() C.double {
	return C.double(d)
}

// C traduce il valore di i nel formato compreso dall'API SDRplay.
func (i integer) C() C.int {
	return C.int(i)
}

// C traduce il valore di bw nel formato compreso dall'API SDRplay.
func (bw B) C() C.mir_sdr_Bw_MHzT {
	return C.mir_sdr_Bw_MHzT(bw)
}

// C traduce il valore di ifm nel formato compreso dall'API SDRplay.
func (ifm IFmode) C() C.mir_sdr_If_kHzT {
	return C.mir_sdr_If_kHzT(ifm)
}

// C traduce il valore di om nel formato compreso dall'API SDRplay.
func (om OffsetMode) C() C.int {
	return C.int(om - 1)
}

// C traduce il valore di olf nel formato compreso dall'API SDRplay.
func (olf LOfrequency) C() C.mir_sdr_LoModeT {
	return C.mir_sdr_LoModeT(olf)
}

// C traduce il valore di df nel formato compreso dall'API SDRplay.
func (df Decimation) C() C.uint {
	return C.uint(df)
}

// C traduce il valore di agc nel formato compreso dall'API SDRplay.
func (agc AGCmode) C() C.mir_sdr_AgcControlT {
	return C.mir_sdr_AgcControlT(agc)
}

// configure permette di configurare la RSP.
func configure(opts ...Option) {
	for _, opt := range opts {
		if opt.apply != nil {
			opt.apply()
		}
	}
}

// band restituisce un valore che rappresenta una delle bande, come definite nel
// tipo mir_sdr_BandT dell'API, in cui ricade la frequenza passata come parametro
// f.
func band(f float64) int {
	switch {
	case f < 12e6:
		return C.mir_sdr_BAND_AM_LO
	case 12e6 <= f && f < 30e6:
		return C.mir_sdr_BAND_AM_MID
	case 30e6 <= f && f < 60e6:
		return C.mir_sdr_BAND_AM_HI
	case 60e6 <= f && f < 120e6:
		return C.mir_sdr_BAND_VHF
	case 120e6 <= f && f < 250e6:
		return C.mir_sdr_BAND_3
	case 250e6 <= f && f < 420e6:
		return C.mir_sdr_BAND_X
	case 420e6 <= f && f < 1000e6:
		return C.mir_sdr_BAND_4_5
	case 1000e6 <= f && f < 2000e6:
		return C.mir_sdr_BAND_L
	default:
		return -1
	}
}
