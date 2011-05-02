// Copyright (C) 2011, Los Alamos National Security, LLC.
// Use of this source code is governed by a BSD-style license.

// This file provides an interface to PAPI's low-level functions.

package papi


/*
#include <stdio.h>
#include <papi.h>

// Because "type" is a keyword in Go, we use some C wrappers to return
// the type field from various structures.
int get_tlb_type(PAPI_mh_tlb_info_t *t) {return t->type;}
int get_cache_type(PAPI_mh_cache_info_t *c) {return c->type;}
*/
import "C"
import "unsafe"
import "os"


// Return the real-time counter's value in clock cycles.
func GetRealCyc() int64 {
	return int64(C.PAPI_get_real_cyc())
}


// Return the real-time counter's value in microseconds.
func GetRealUsec() int64 {
	return int64(C.PAPI_get_real_usec())
}


// Return the virtual-time counter's value in clock cycles.
func GetVirtCyc() int64 {
	return int64(C.PAPI_get_virt_cyc())
}


// Return the virtual-time counter's value in microseconds.
func GetVirtUsec() int64 {
	return int64(C.PAPI_get_virt_usec())
}


// Return the executable's address-space information.
func GetExecutableInfo() ProgramInfo {
	cinfo := C.PAPI_get_executable_info()
	if cinfo == nil {
		// I can't imagine this ever happening, but we should
		// do something just in case.
		panic("PAPI_get_executable_info() failed unexpectedly")
	}
	addrInfo := cinfo.address_info
	return ProgramInfo{
		FullName: C.GoString(&cinfo.fullname[0]),
		AddressInfo: AddressMap{
			Name:      C.GoString(&addrInfo.name[0]),
			TextStart: uintptr(unsafe.Pointer(addrInfo.text_start)),
			TextEnd:   uintptr(unsafe.Pointer(addrInfo.text_end)),
			DataStart: uintptr(unsafe.Pointer(addrInfo.data_start)),
			DataEnd:   uintptr(unsafe.Pointer(addrInfo.data_end)),
			BssStart:  uintptr(unsafe.Pointer(addrInfo.bss_start)),
			BssEnd:    uintptr(unsafe.Pointer(addrInfo.bss_end))}}
}


// Acquire and return all sorts of information about the underlying hardware.
func GetHardwareInfo() HardwareInfo {
	hw := C.PAPI_get_hardware_info()
	maxLevels := int(C.PAPI_MH_MAX_LEVELS)

	// Describe all levels of the memory hierarchy.
	mh := make([]MHLevelInfo, hw.mem_hierarchy.levels)
	for level, _ := range mh {
		cLevel := hw.mem_hierarchy.level[level]

		// Populate the TLB information.
		tlbData := make([]TLBInfo, maxLevels)
		var validTLBLevels int
		for i, _ := range tlbData {
			ctlb := cLevel.tlb[i]
			tlbData[i].Type = MHAttrs(C.get_tlb_type(&ctlb))
			if tlbData[i].Type == MH_TYPE_EMPTY {
				break
			}
			tlbData[i].NumEntries = int32(ctlb.num_entries)
			tlbData[i].PageSize = int32(ctlb.page_size)
			tlbData[i].Associativity = int32(ctlb.associativity)
			validTLBLevels++
		}
		mh[level].TLB = tlbData[0:validTLBLevels]

		// Populate the cache information.
		cacheData := make([]CacheInfo, maxLevels)
		var validCacheLevels int
		for i, _ := range cacheData {
			ccache := cLevel.cache[i]
			cacheData[i].Type = MHAttrs(C.get_cache_type(&ccache))
			if cacheData[i].Type == MH_TYPE_EMPTY {
				break
			}
			cacheData[i].Size = int32(ccache.size)
			cacheData[i].LineSize = int32(ccache.line_size)
			cacheData[i].NumLines = int32(ccache.num_lines)
			cacheData[i].Associativity = int32(ccache.associativity)
			validCacheLevels++
		}
		mh[level].Cache = cacheData[0:validCacheLevels]
	}

	// Populate and return the set of available hardware information.
	return HardwareInfo{
		CPUs:          int32(hw.ncpu),
		Threads:       int32(hw.threads),
		Cores:         int32(hw.cores),
		Sockets:       int32(hw.sockets),
		NUMANodes:     int32(hw.nnodes),
		TotalCPUs:     int32(hw.totalcpus),
		Vendor:        int32(hw.vendor),
		VendorName:    C.GoString(&hw.vendor_string[0]),
		Model:         int32(hw.model),
		ModelName:     C.GoString(&hw.model_string[0]),
		Revision:      float32(hw.revision),
		CPUIDFamily:   int32(hw.cpuid_family),
		CPUIDModel:    int32(hw.cpuid_model),
		CPUIDStepping: int32(hw.cpuid_stepping),
		MHz:           float32(hw.mhz),
		ClockMHz:      int32(hw.clock_mhz),
		MemHierarchy:  mh}
}


// Acquire and return all sorts of information about the current
// process's dynamic memory usage.  In addition to returning an
// overall error code, GetDynMemInfo() can also return an Errno cast
// to an int64 for any individual field.  To check for that case, note
// that all errors are represented as negative Errno values.
func GetDynMemInfo() (dmem DynMemInfo, err os.Error) {
	var c_dmem C.PAPI_dmem_info_t
	if err = Errno(C.PAPI_get_dmem_info(&c_dmem)); err != OK {
		return
	}
	dmem = DynMemInfo{
		Peak:          int64(c_dmem.peak),
		Size:          int64(c_dmem.size),
		Resident:      int64(c_dmem.resident),
		HighWaterMark: int64(c_dmem.high_water_mark),
		Shared:        int64(c_dmem.shared),
		Text:          int64(c_dmem.text),
		Library:       int64(c_dmem.library),
		Heap:          int64(c_dmem.heap),
		Locked:        int64(c_dmem.locked),
		Stack:         int64(c_dmem.stack),
		PageSize:      int64(c_dmem.pagesize),
		PTE:           int64(c_dmem.pte)}
	err = nil
	return
}
