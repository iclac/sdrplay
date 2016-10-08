/*
   sdrplay is a Go package that enables to use the RSP (by SDRplay) in a Go program.
   Copyright (C) 2016 Claudio Carraro carraro.claudio@gmail.com

   See the COPYING file to GPLv2 license details.
*/

package sdrplay

import "errors"

type (
	// Tuner è l'interfaccia che descrive un sintonizzatore radio.
	Tuner interface {
		// Tune permette di sintonizzare una desiderata frequenza. In particolare
		// imposta come frequenza centrale del sintonizzatore interno alla RSP il
		// valore frequency espresso in Hz.
		Tune(frequency float64) error
	}

	// Amplifier è l'interfaccia che rappresenta un amplificatore.
	Amplifier interface {
		// Gain permette di impostare un valore di guadagno. In particolare, da
		// quanto descritto in http://www.sdrplay.com/docs/SDRplay_AGC_technote_r2p2.pdf,
		// l'API RSP permette di impostare dei valori di gain reduction rispetto
		// al valore massimo di guadagno possibile nella RSP.
		Gain(reduction int) error
	}

	// Receiver è l'interfaccia che descrive un semplice ricevitore radio.
	Receiver interface {
		Tuner
		Amplifier
		SetUp(opts ...Option) error
	}

	// Connector è l'interfaccia che descrive un connettore, ossia il mezzo
	// attraverso il quale si possono propagare i segnali prodotti dalla relativa
	// sorgente.
	Connector interface {
		// Propagate permette alla sorgente di un segnale, di cui Connector è il
		// connettore verso i possibili utilizzatori, di propagare il segnale
		// stesso. In particolare il segnale propagato è la rappresentazione in
		// banda base del segnale ricevuto dalla RSP. Tale segnale ha le due
		// componenti in fase (I) e in quadratura (Q) tipiche di tale
		// rappresentazione. Queste due componenti sono di tipo []int16 perchè
		// quanto propagato è un frame di campioni castati al tipo Go più vicino
		// allo short del C generato dalla RSP.
		Propagate(I []int16, Q []int16)
	}

	// Option rappresenta un'opzione di configurazione di RSP.
	Option struct {
		apply func()
	}
)

var (
	// DeactivatedReceiverError indica che il ricevitore, sul quale è stata
	// invocata l'operazione che ha prodotto tale errore, è stato disattivato a
	// causa della creazione di un nuovo ricevitore operata dalla funzione RSP.
	DeactivatedReceiverError = errors.New("Deactivated Receiver Error")

	// UnpluggedConnectorError indica che non è stato fornito un connettore alla
	// funzione RSP.
	UnpluggedConnectorError = errors.New("Unplugged Connector Error")
)

// RSP permette di ottenere un ricevitore con le caratteristiche desiderate (opts)
// fornendo la rappresentazione in banda base del segnale desiderato al Connector
// fornito.
// Ad ogni invocazione, se presente, il precedente receiver verrà disattivato ed
// ogni suo metodo fornirà l'errore DeactivatedReceiverError.
// Il baseband connector deve essere non nil altrimenti viene restituito l'errore
// UnpluggedConnectorError. Le opzioni opts sono facoltative, se non presenti
// verrà usata una configurazione di default.
func RSP(baseband Connector, opts ...Option) (Receiver, error) {
	if baseband == nil {
		return nil, UnpluggedConnectorError
	}

	if rx != nil {
		e := rx.uninit()
		if e != nil {
			return nil, e
		}

		// Si disattiva il precedente ricevitore.
		rx.baseband = nil
	}

	newRadio()

	rsp = features{}

	configure(fm102MHz...)
	configure(opts...)

	rx.feat = rsp
	rx.baseband = baseband

	ie := rx.init()

	return rx, ie
}

// B enumera tutte le larghezze di banda ammesse.
type B int

const (
	// BW200 indica una larghezza di banda pari a 200kHz.
	BW200 B = 200
	// BW300 indica una larghezza di banda pari a 300kHz.
	BW300 B = 300
	// BW600 indica una larghezza di banda pari a 600kHz.
	BW600 B = 600
	// BW1536 indica una larghezza di bada pari a 1536kHz.
	BW1536 B = 1536
	// BW5000 indica una larghezza di banda pari a 5MHz.
	BW5000 B = 5000
	// BW6000 indica una larghezza di banda pari a 6MHz.
	BW6000 B = 6000
	// BW7000 indica una larghezza di banda pari a 7MHz.
	BW7000 B = 7000
	// BW8000 indica una larghezza di banda pari a 8MHz.
	BW8000 B = 8000
)

// Bandwidth permette di impostare la larghezza di banda.
func Bandwidth(bw B) Option {
	return Option{
		apply: func() {
			rsp.BW = bw
		},
	}
}

// IFmode enumera tutti i valori ammessi di frequenza intermedia.
type IFmode int

const (
	// IFzero indica un valore di IF pari a 0Hz.
	IFzero IFmode = 0
	// IF450 indica un valore di IF pari a 450kHz.
	IF450 IFmode = 450
	// IF1620 indica un valore di IF pari a 1620kHz.
	IF1620 IFmode = 1620
	// IF2048 indica un valore di IF pari a 2048kHz.
	IF2048 IFmode = 2048
)

// IF permette di impostare il valore della frequenza intermedia.
func IF(ifreq IFmode) Option {
	return Option{
		apply: func() {
			rsp.IF = ifreq
		},
	}
}

// FS permette di impostare la frequenza di campionamento espressa in Hz.
func FS(hz float64) Option {
	return Option{
		apply: func() {
			rsp.FS = double(hz)
		},
	}
}

// IQimbalance permette di abilitare o meno la correzione del IQ imbalance.
func IQimbalance(enabled bool) Option {
	return Option{
		apply: func() {
			rsp.IQimbalance = enable(enabled)
		},
	}
}

// DCoffset permette di abilitare o meno la correzione del offset DC.
func DCoffset(enabled bool) Option {
	return Option{
		apply: func() {
			rsp.DCoffset = enable(enabled)
		},
	}
}

// OffsetMode descrive il metodo di correzione dell'offset DC.
type OffsetMode int

const (
	// None è il valore di default che indica che nessuna impostazione è stata
	// eseguita. Essendo tale valore pari a 0, quelli a seguire sono aumentati
	// di una unità rispetto ai valori presenti nella API SDRplay. Usando la
	// funzione mir_sdr_SetDcMode al parametro OffsetMode, se diverso da 0,
	// dovrà essere sottratto 1: questo è automaticamente fatto dal metodo C.
	None OffsetMode = iota
	// Static non esegue nessuna correzione dell'offset DC.
	Static
	// Periodic6ms applica la correzione periodicamente ogni 6ms
	Periodic6ms
	// Periodic12ms applica la correzione periodicamente ogni 12ms
	Periodic12ms
	// Periodic24ms applica la correzione periodicamente ogni 24ms
	Periodic24ms
	// OneShot applica la correzione ogni volta che viene aggiornato il guadagno
	OneShot
	// Continuous applica continuamente la correzione.
	Continuous
)

// DCmode imposta il metodo di correzione dell'offset DC del ricevitore.
func DCmode(mode OffsetMode) Option {
	return Option{
		apply: func() {
			rsp.DCmode = mode
		},
	}
}

// DCtrakTime imposta il periodo di tempo nel quale viene monitorato il DC offset
// quando il DC mode è impostato a OneShot.
// Valori ammessi nell'intervallo 1-63 i quali corrispondono ad una durata di
// monitoraggio di 3*trackTime us.
// Il valore trackTime passato viene riportato all'intervallo ammesso nel seguente
// modo:
//    DCtrackTime =  1 se trackTime < 1
//    DCtrackTime = 63 se trackTime > 63
func DCtrackTime(trackTime int) Option {
	tt := 0
	switch {
	case trackTime < 1:
		tt = 1
	case trackTime > 63:
		tt = 63
	default:
		tt = trackTime
	}

	return Option{
		apply: func() {
			rsp.DCTrakTime = integer(tt)
		},
	}
}

// LOppm imposta il fattore di correzione per tener conto del offset della
// frequenza nominale dell'oscillatore locale.
// Il valore ppm verrà castato al tipo double dell'API C.
func LOppm(ppm float64) Option {
	return Option{
		apply: func() {
			rsp.LOppm = double(ppm)
		},
	}
}

// LOfrequency specifica la frequenza dell'OL (oscillatore locale) del up-converter
// usato quando si sintonizzano frequenze inferiori ai 60MHz oppure per quelle
// comprese tra 250 e 420 MHz.
type LOfrequency int

const (
	// LOundefined è il valore di default che indica che nessuna impostazione è
	// stata eseguita.
	LOundefined LOfrequency = iota
	// LOauto seleziona automaticamente la frequenza del OL fornendo un'appropriata
	// copertura in tutta la gamma di frequenze RF.
	LOauto
	// LO120MHz imposta la frequenza del OL a 120MHz ottenendo la copertura per
	// il range di frequenze 370-420MHz.
	LO120MHz
	// LO144MHz imposta la frequenza del OL a 144MHz ottenendo la copertura per
	// il range di frequenze 250-255MHz e 400-420MHz.
	LO144MHz
	// LO168MHz imposta la frequenza del OL a 168MHz ottenendo la copertura per
	// il range di frequenze 250-265MHz.
	LO168MHz
)

// LOmode imposta la frequenza del up-converter, oppure indica all'API di determinare
// il valore più appropriato della frequenza del OL.
func LOmode(loMode LOfrequency) Option {
	return Option{
		apply: func() {
			rsp.LOmode = loMode
		},
	}
}

// Decimation enumera il fattore di decimazione.
type Decimation int

const (
	// Factor0 indica nessuna decimazione.
	Factor0 Decimation = 0
	// Factor2 indica un fattore di decimazione pari a 2.
	Factor2 Decimation = 1 << (iota + 1)
	// Factor4 indica un fattore di decimazione pari a 4.
	Factor4
	// Factor8 indica un fattore di decimazione pari a 8.
	Factor8
	// Factor16 indica un fattore di decimazione pari a 16.
	Factor16
	// Factor32 indica un fattore di decimazione pari a 32.
	Factor32
	// Factor64 indica un fattore di decimazione pari a 64.
	Factor64
)

// Decimate permette di abilitare o meno la decimazione e specifica il fattore di
// decimazione.
func Decimate(enabled bool, factor Decimation) Option {
	return Option{
		apply: func() {
			rsp.Decimate = enable(enabled)
			rsp.Factor = factor
		},
	}
}

// LNA permette di abilitare o meno l'amplificatore a basso rumore.
func LNA(enabled bool) Option {
	return Option{
		apply: func() {
			rsp.LNA = enable(enabled)
		},
	}
}

//AGCmode specifica il modo di controllo del AGC.
type AGCmode int

const (
	// Disable disabilita l'AGC.
	Disable AGCmode = iota
	// AGC100Hz abilita l'AGC con loop a 100Hz.
	AGC100Hz
	// AGC50Hz abilita l'AGC con loop a 50Hz.
	AGC50Hz
	// AGC5Hz abilita l'AGC con loop a 5Hz.
	AGC5Hz
)

// AGC permette di abilitare o meno il controllo automatico del guadagne e di
// impostare il valore desiderato dell'intensità del segnale RF espresso in dBFS.
// (dBFS è un valore di misura di potenza di un segnale relativo al fondo scala,
// quindi il valore massimo è pari a 0dBFS. Quindi il parametro passato alla
// funzione deve essere minore, o al più uguale, a 0).
func AGC(mode AGCmode, dBFS int) Option {
	return Option{
		apply: func() {
			rsp.AGC = mode
			rsp.DBFS = integer(dBFS)
		},
	}
}

// InitialGR imposta il valore iniziale di gain reduction in dB.
func InitialGR(dB int) Option {
	return Option{
		apply: func() {
			rsp.InitialGR = integer(dB)
		},
	}
}

// InitialRF imposta il valore iniziale della frequenza sintonizzata. Il valore
// frequency viene considerato espresso in MHz.
func InitialRF(frequency float64) Option {
	return Option{
		apply: func() {
			rsp.InitialRF = double(frequency)
		},
	}
}

// Debug permette di abilitare o meno i messaggi di debug dalla libreria SDRplay.
func Debug(enabled bool) Option {
	return Option{
		apply: func() {
			rsp.Debug = enable(enabled)
		},
	}
}
