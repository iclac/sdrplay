/*
   sdrplay is a Go package that enables to use the RSP (by SDRplay) in a Go program.
   Copyright (C) 2016 Claudio Carraro carraro.claudio@gmail.com

   See the COPYING file to GPLv2 license details.
*/

// sdrplay è un package che permette di usare la RSP, la SDR di SDRplay, in un
// programma Go. Il package maschera però l'API originale, comunque usata
// attraverso cgo, ma cerca di esporne una più immediata.
package sdrplay
