#include <windows.h>
#include <psapi.h>
#include <stdio.h>

/* Based on https://github.com/hlldz/RefleXXion (95c0ee1cce2adfff1b74f2d1f84b7355a5befc60)
   No license 
   Some C compatibility patches
*/

#define NT_SUCCESS(Status) ((NTSTATUS)(Status) >= 0)
#define NtCurrentProcess() ( (HANDLE)(LONG_PTR) -1 )

#define MAX_SYSCALL_STUB_SIZE 64

#define OBJ_CASE_INSENSITIVE 0x40

#define InitializeObjectAttributes( p, n, a, r, s ) {   \
    (p)->Length = sizeof( OBJECT_ATTRIBUTES );          \
    (p)->RootDirectory = r;                             \
    (p)->Attributes = a;                                \
    (p)->ObjectName = n;                                \
    (p)->SecurityDescriptor = s;                        \
    (p)->SecurityQualityOfService = NULL;               \
    }

typedef void* PRTL_USER_PROCESS_PARAMETERS;

typedef void* PPS_POST_PROCESS_INIT_ROUTINE;

typedef struct _UNICODE_STRING {
	USHORT Length;
	USHORT MaximumLength;
	PWSTR  Buffer;
} UNICODE_STRING, * RX_UNICODE_STRING;

typedef struct _PEB_LDR_DATA {
	DWORD dwLength;
	DWORD dwInitialized;
	LPVOID lpSsHandle;
	LIST_ENTRY InLoadOrderModuleList;
	LIST_ENTRY InMemoryOrderModuleList;
	LIST_ENTRY InInitializationOrderModuleList;
	LPVOID lpEntryInProgress;
} PEB_LDR_DATA, * RX_PEB_LDR_DATA;

typedef struct _PEB {
	BYTE Reserved1[2];
	BYTE BeingDebugged;
	BYTE Reserved2[1];
	PVOID Reserved3[2];
	RX_PEB_LDR_DATA Ldr;
	PRTL_USER_PROCESS_PARAMETERS ProcessParameters;
	PVOID Reserved4[3];
	PVOID AtlThunkSListPtr;
	PVOID Reserved5;
	ULONG Reserved6;
	PVOID Reserved7;
	ULONG Reserved8;
	ULONG AtlThunkSListPtr32;
	PVOID Reserved9[45];
	BYTE Reserved10[96];
	PPS_POST_PROCESS_INIT_ROUTINE PostProcessInitRoutine;
	BYTE Reserved11[128];
	PVOID Reserved12[1];
	ULONG SessionId;
} PEB, * RX_PEB;

typedef struct {
	LIST_ENTRY InLoadOrderLinks;
	LIST_ENTRY InMemoryOrderModuleList;
	LIST_ENTRY InInitializationOrderModuleList;
	PVOID DllBase;
	PVOID EntryPoint;
	ULONG SizeOfImage;
	UNICODE_STRING FullDllName;
	UNICODE_STRING BaseDllName;
	ULONG Flags;
	SHORT LoadCount;
	SHORT TlsIndex;
	LIST_ENTRY HashTableEntry;
	ULONG TimeDateStamp;
} LDR_DATA_TABLE_ENTRY, * RX_LDR_DATA_TABLE_ENTRY;

typedef struct _OBJECT_ATTRIBUTES {
	ULONG Length;
	HANDLE RootDirectory;
	RX_UNICODE_STRING ObjectName;
	ULONG Attributes;
	PVOID SecurityDescriptor;        // Points to type SECURITY_DESCRIPTOR
	PVOID SecurityQualityOfService;  // Points to type SECURITY_QUALITY_OF_SERVICE

} OBJECT_ATTRIBUTES, * RX_OBJECT_ATTRIBUTES;

typedef struct _IO_STATUS_BLOCK {
	union {
		NTSTATUS Status;
		PVOID Pointer;
	};
	ULONG_PTR Information;
} IO_STATUS_BLOCK, * RX_IO_STATUS_BLOCK;

// NtOpenFile
typedef NTSYSAPI NTSTATUS(NTAPI* RX_NtOpenFile)(PHANDLE, ACCESS_MASK, RX_OBJECT_ATTRIBUTES, RX_IO_STATUS_BLOCK, ULONG, ULONG);

// NtCreateSection
typedef NTSTATUS(NTAPI* RX_NtCreateSection)(PHANDLE, ACCESS_MASK, RX_OBJECT_ATTRIBUTES, PLARGE_INTEGER, ULONG, ULONG, HANDLE);

// NtOpenSection
typedef NTSTATUS(NTAPI* RX_NtOpenSection)(HANDLE*, ACCESS_MASK, OBJECT_ATTRIBUTES*);

// NtMapViewOfSection
typedef NTSTATUS(NTAPI* RX_NtMapViewOfSection)(HANDLE, HANDLE, PVOID*, ULONG_PTR, SIZE_T, PLARGE_INTEGER, PSIZE_T, DWORD, ULONG, ULONG);

// NtProtectVirtualMemory
typedef NTSTATUS(NTAPI* _NtProtectVirtualMemory)(HANDLE, PVOID, PSIZE_T, ULONG, PULONG);

// NtUnmapViewOfSection
typedef NTSTATUS(NTAPI* _NtUnmapViewOfSection)(HANDLE, PVOID);

// RtlInitUnicodeString
typedef VOID(NTAPI* _RtlInitUnicodeString)(RX_UNICODE_STRING, PCWSTR);

// Needed for Technique - 1 (Reading NTDLL From Disk)
RX_NtOpenFile RxNtOpenFile = NULL;
RX_NtCreateSection RxNtCreateSection = NULL;

// Needed for Technique - 2 (Reading NTDLL From KnownDlls)
RX_NtOpenSection RxNtOpenSection = NULL;

// Needed for Both Technique - 1 and Technique - 2
RX_NtMapViewOfSection RxNtMapViewOfSection = NULL;

void debug(const char *fmt, ...)
{
#ifdef DEBUG
  va_list ap;
  va_start(ap, fmt);
  vfprintf(stderr, fmt, ap);
  va_end(ap);
#endif
}


ULONG_PTR BuildSyscallStub(ULONG_PTR pStubRegion, DWORD dwSyscallNo) {
	BYTE bSyscallStub[] = {
			0x4c, 0x8b, 0xd1,				// mov     r10,rcx
			0xb8, 0x00, 0x00, 0x00, 0x00,	// mov     eax,xxx
			0x0f, 0x05,						// syscall
			0xc3							// ret
	};

	memcpy((PBYTE)pStubRegion, bSyscallStub, sizeof(bSyscallStub));
	*(DWORD*)(pStubRegion + 4) = dwSyscallNo;

	return pStubRegion;
}

int InitSyscallsFromLdrpThunkSignature() {
	debug("[*] Parsing LdrpThunkSignature For Clean Syscalls.\n\n");

	RX_PEB pPEB = (RX_PEB)__readgsqword(0x60);
	RX_PEB_LDR_DATA pPEBLdr = pPEB->Ldr;
	RX_LDR_DATA_TABLE_ENTRY pLdeNTDLL = NULL;

	for (RX_LDR_DATA_TABLE_ENTRY pLdeTmp = (RX_LDR_DATA_TABLE_ENTRY)pPEBLdr->InLoadOrderModuleList.Flink; pLdeTmp->DllBase != NULL; pLdeTmp = (RX_LDR_DATA_TABLE_ENTRY)pLdeTmp->InLoadOrderLinks.Flink) {
		if (_wcsnicmp(pLdeTmp->BaseDllName.Buffer, L"ntdll.dll", 9) == 0) {
			// We Detect NTDLL
			pLdeNTDLL = pLdeTmp;
			break;
		}
	}

	if (pLdeNTDLL == NULL) {
		debug("[!] Cannot find NTDLL.\n");
		return 0;
	}

	PIMAGE_NT_HEADERS ImageNtHeaders = (PIMAGE_NT_HEADERS)((ULONG_PTR)pLdeNTDLL->DllBase + ((PIMAGE_DOS_HEADER)pLdeNTDLL->DllBase)->e_lfanew);
	PIMAGE_SECTION_HEADER SectionHeader = (PIMAGE_SECTION_HEADER)((ULONG_PTR)&ImageNtHeaders->OptionalHeader + ImageNtHeaders->FileHeader.SizeOfOptionalHeader);

	ULONG_PTR DataSectionAddress = (ULONG_PTR)NULL;
	DWORD DataSectionSize;

	for (WORD i = 0; i < ImageNtHeaders->FileHeader.NumberOfSections; i++) {
		if (!strcmp((char*)SectionHeader[i].Name, ".data")) {
			DataSectionAddress = (ULONG_PTR)pLdeNTDLL->DllBase + SectionHeader[i].VirtualAddress;
			DataSectionSize = SectionHeader[i].Misc.VirtualSize;
			break;
		}
	}

	if (!DataSectionAddress || DataSectionSize < 16 * 5) {
		return 0;
	}

	// NtOpenFile Syscall Number
	DWORD dwNtOpenFile = 0;

	// NtCreateSection Syscall Number
	DWORD dwNtCreateSection = 0;

	// NtOpenSection Syscall Number
	DWORD dwNtOpenSection = 0;

	// NtMapViewOfSection Syscall Number
	DWORD dwNtMapViewOfSection = 0;

	for (UINT uiOffset = 0; uiOffset < DataSectionSize - (16 * 5); uiOffset++) {
		if (*(DWORD*)(DataSectionAddress + uiOffset) == 0xb8d18b4c &&
			*(DWORD*)(DataSectionAddress + uiOffset + 16) == 0xb8d18b4c &&
			*(DWORD*)(DataSectionAddress + uiOffset + 32) == 0xb8d18b4c &&
			*(DWORD*)(DataSectionAddress + uiOffset + 48) == 0xb8d18b4c &&
			*(DWORD*)(DataSectionAddress + uiOffset + 64) == 0xb8d18b4c) {

			dwNtOpenFile = *(DWORD*)(DataSectionAddress + uiOffset + 4); // Needed for Technique - 1 (Reading NTDLL From Disk)

			dwNtCreateSection = *(DWORD*)(DataSectionAddress + uiOffset + 16 + 4); // Needed for Technique - 1 (Reading NTDLL From Disk)

			dwNtOpenSection = *(DWORD*)(DataSectionAddress + uiOffset + 48 + 4); // Needed for Technique - 2 (Reading NTDLL From KnownDlls)

			dwNtMapViewOfSection = *(DWORD*)(DataSectionAddress + uiOffset + 64 + 4); // Needed for Both Technique - 1 and Technique - 2

			break;
		}
	}

	// A little check to see if we have successfully parsed syscall numbers
	if (!dwNtMapViewOfSection) {

		return 0;

	}

	debug("[*] NtOpenFile syscall number..............: %02x\n", dwNtOpenFile);

	debug("[*] NtCreateSection syscall number.........: %02x\n", dwNtCreateSection);

	debug("[*] NtOpenSection syscall number...........: %02x\n", dwNtOpenSection);

	debug("[*] NtMapViewOfSection syscall number......: %02x\n\n", dwNtMapViewOfSection);

	// Create RX memory region for syscalls stub
	ULONG_PTR pSyscallRegion = (ULONG_PTR)VirtualAlloc(NULL, 4 * MAX_SYSCALL_STUB_SIZE, MEM_RESERVE | MEM_COMMIT, PAGE_READWRITE);

	if (!pSyscallRegion) {

		debug("[!] Cannot allocate memory for syscals stubs.\n");

		return 0;

	}

	// Create NtOpenFile clean syscall memory region with stub
	RxNtOpenFile = (RX_NtOpenFile)BuildSyscallStub(pSyscallRegion, dwNtOpenFile);

	// Create NtCreateSection clean syscall memory region with stub
	RxNtCreateSection = (RX_NtCreateSection)BuildSyscallStub(pSyscallRegion + MAX_SYSCALL_STUB_SIZE, dwNtCreateSection);

	// Create NtOpenSection clean syscall memory region with stub
	RxNtOpenSection = (RX_NtOpenSection)BuildSyscallStub(pSyscallRegion + (2 * MAX_SYSCALL_STUB_SIZE), dwNtOpenSection);

	// Create NtMapViewOfSection clean syscall memory region with stub
	RxNtMapViewOfSection = (RX_NtMapViewOfSection)BuildSyscallStub(pSyscallRegion + (3 * MAX_SYSCALL_STUB_SIZE), dwNtMapViewOfSection);

	// Modify the syscall memory region to RX
	DWORD dwOldProtection;
	BOOL bStatus = VirtualProtect((LPVOID)pSyscallRegion, 4 * MAX_SYSCALL_STUB_SIZE, PAGE_EXECUTE_READ, &dwOldProtection);

	return 1;
}

// Custom x64 GetProcAddress Implementation for NtProtectVirtualMemory executions from https://revers.engineering/custom-getprocaddress-and-getmodulehandle-implementation-x64/
uintptr_t CustomGetProcAddress(void* hModule, const char* wAPIName) {
	unsigned char* lpBase = (unsigned char*)(hModule);
	PIMAGE_DOS_HEADER idhDosHeader = (PIMAGE_DOS_HEADER)(lpBase);

	if (idhDosHeader->e_magic == 0x5A4D) {
		PIMAGE_NT_HEADERS inhNtHeader = (PIMAGE_NT_HEADERS)(lpBase + idhDosHeader->e_lfanew);

		if (inhNtHeader->Signature == 0x4550) {
			PIMAGE_EXPORT_DIRECTORY iedExportDirectory = (PIMAGE_EXPORT_DIRECTORY)(lpBase + inhNtHeader->OptionalHeader.DataDirectory[IMAGE_DIRECTORY_ENTRY_EXPORT].VirtualAddress);

			for (register unsigned int uiIter = 0; uiIter < iedExportDirectory->NumberOfNames; ++uiIter) {
				for (register unsigned int uiIter = 0; uiIter < iedExportDirectory->NumberOfNames; ++uiIter) {
					char* szNames = (char*)(lpBase + ((unsigned long*)(lpBase + iedExportDirectory->AddressOfNames))[uiIter]);

					if (strcmp(szNames, wAPIName) == 0) {
						unsigned short usOrdinal = (unsigned short)((unsigned short*)(lpBase + iedExportDirectory->AddressOfNameOrdinals))[uiIter];
						uintptr_t ret = (uintptr_t)(lpBase + ((unsigned long*)(lpBase + iedExportDirectory->AddressOfFunctions))[usOrdinal]);
						return ret;
					}
				}
			}
		}
	}

	return (uintptr_t) NULL;
}

// Technique - 1
// Reads NTDLL From Disk and Clean
int Technique1() {

	debug("[*] Using Technique-1, Reads NTDLL From Disk and Clean.\n\n");

	NTSTATUS ntStatus;

	// Get handle to loaded/hooked NTDLL
	HMODULE hHookedNtdll = GetModuleHandleA("ntdll.dll");

	if (hHookedNtdll == NULL) {
		debug("[-] GetModuleHandleA error: %d\n", GetLastError());
		return 0;
	}

	OBJECT_ATTRIBUTES ObjectAttributes = { 0 };
	UNICODE_STRING ObjectPath = { 0 };
	IO_STATUS_BLOCK IoStatusBlock = { 0 };
	_RtlInitUnicodeString RtlInitUnicodeString = (_RtlInitUnicodeString)GetProcAddress(hHookedNtdll, "RtlInitUnicodeString");
	if (RtlInitUnicodeString == NULL) {
		return 0;
	}
	RtlInitUnicodeString(&ObjectPath, L"\\??\\C:\\Windows\\System32\\ntdll.dll");

	InitializeObjectAttributes(&ObjectAttributes, &ObjectPath, OBJ_CASE_INSENSITIVE, NULL, NULL);

	HANDLE hFile = NULL;

	ntStatus = RxNtOpenFile(&hFile, FILE_READ_DATA | GENERIC_READ, &ObjectAttributes, &IoStatusBlock, FILE_SHARE_READ, (ULONG)0);

	if (!NT_SUCCESS(ntStatus)) {
		debug("[-] NtOpenFile error.\n");
		CloseHandle(hFile);
		return 0;
	}

	debug("[*] Clean NTDLL Handle Address.............: 0x%p\n", hFile);

	HANDLE hSection = NULL;

	ntStatus = RxNtCreateSection(&hSection, STANDARD_RIGHTS_REQUIRED | SECTION_MAP_READ | SECTION_QUERY, NULL, NULL, PAGE_READONLY, SEC_IMAGE, hFile);

	if (!NT_SUCCESS(ntStatus)) {
		debug("[-] NtCreateSection error.\n");
		CloseHandle(hSection);
		return 0;
	}

	debug("[*] Clean Section Handle Address...........: 0x%p\n\n", hSection);

	LPVOID pCleanNtdll = NULL;
	SIZE_T sztViewSize = 0;

	ntStatus = RxNtMapViewOfSection(hSection, NtCurrentProcess(), &pCleanNtdll, (ULONG_PTR)NULL, (SIZE_T)NULL, NULL, &sztViewSize, 1, 0, PAGE_READONLY);

	if (!NT_SUCCESS(ntStatus)) {
		debug("[-] NtMapViewOfSection error.\n");
		CloseHandle(hSection);
		return 0;
	}

	debug("[*] Clean NTDLL Base Address...............: 0x%p\n", pCleanNtdll);

	MODULEINFO miHookedNtdll = {};

	if (GetModuleInformation(NtCurrentProcess(), hHookedNtdll, &miHookedNtdll, sizeof(miHookedNtdll)) == 0) {
		debug("[-] GetModuleInformation error: %d\n", GetLastError());
		return 0;
	}

	// Get base address of hooked NTDLL from MODULEINFO struct
	LPVOID pHookedBaseAddress = (LPVOID)miHookedNtdll.lpBaseOfDll;
	debug("[*] Hooked NTDLL Base Address..............: 0x%p\n\n", pHookedBaseAddress);

	// Get hooked NTDLL DOS header
	PIMAGE_DOS_HEADER pHookedDosHeader = (PIMAGE_DOS_HEADER)pHookedBaseAddress;

	// Get hooked NTDLL NT header
	PIMAGE_NT_HEADERS pHookedNtHeader = (PIMAGE_NT_HEADERS)((DWORD_PTR)pHookedBaseAddress + pHookedDosHeader->e_lfanew);

	for (SIZE_T i = 0; i < pHookedNtHeader->FileHeader.NumberOfSections; i++) {

		// Get PE section header
		PIMAGE_SECTION_HEADER pHookedSectionHeader = (PIMAGE_SECTION_HEADER)((DWORD_PTR)IMAGE_FIRST_SECTION(pHookedNtHeader) + ((DWORD_PTR)IMAGE_SIZEOF_SECTION_HEADER * i));

		// Get section name
		LPSTR szHookedSectionName = (LPSTR)pHookedSectionHeader->Name;

		// We found .TEXT section
		if (!strcmp(szHookedSectionName, ".text")) {

			// Get start address of hooked .TEXT section
			LPVOID pHookedTextSectionAddress = (LPVOID)((DWORD_PTR)pHookedBaseAddress + (DWORD_PTR)pHookedSectionHeader->VirtualAddress);

			// Get start address of clean .TEXT section
			LPVOID pCleanTextStartAddress = (LPVOID)((DWORD_PTR)pCleanNtdll + (DWORD_PTR)pHookedSectionHeader->VirtualAddress);

			// Get size of .TEXT section
			SIZE_T sztTextSectionSize = pHookedSectionHeader->Misc.VirtualSize;

			debug("[*] Hooked NTDLL .TEXT Section VA..........: 0x%p\n", pHookedTextSectionAddress);
			debug("[*] Clean NTDLL .TEXT Section VA...........: 0x%p\n\n", pCleanTextStartAddress);
			debug("[*] Size of .TEXT Section..................: %zd\n", sztTextSectionSize);

			// Change original page protection of hooked NTDLL to RWX
			LPVOID lpBaseAddress = pHookedTextSectionAddress;
			SIZE_T uSize = sztTextSectionSize;

			_NtProtectVirtualMemory NtProtectVirtualMemory = (_NtProtectVirtualMemory)CustomGetProcAddress(pCleanNtdll, "NtProtectVirtualMemory");

			ULONG oldProtection;
			ntStatus = NtProtectVirtualMemory(NtCurrentProcess(), &lpBaseAddress, &uSize, PAGE_EXECUTE_READWRITE, &oldProtection);

			if (!NT_SUCCESS(ntStatus)) {
				debug("[-] NtProtectVirtualMemory - 1: Error.\n");
				return 0;
			}

			// Copy .TEXT section of clean NTDLL into hooked NTDLL .TEXT section
			memcpy(pHookedTextSectionAddress, pCleanTextStartAddress, sztTextSectionSize);

			// Revert back to original page protections of overwritten NTDLL .TEXT section
			ntStatus = NtProtectVirtualMemory(NtCurrentProcess(), &lpBaseAddress, &uSize, oldProtection, &oldProtection);
			if (!NT_SUCCESS(ntStatus)) {
				debug("[-] NtProtectVirtualMemory - 2: Error.\n");
				return 0;
			}

			debug("\n[+] NTDLL is cleaned. Closing handles...\n\n");

			break;

		}

	}

	// Unmap the local section
	_NtUnmapViewOfSection NtUnmapViewOfSection = (_NtUnmapViewOfSection)GetProcAddress(hHookedNtdll, "NtUnmapViewOfSection");
	ntStatus = NtUnmapViewOfSection(NtCurrentProcess(), pCleanNtdll);
	if (!NT_SUCCESS(ntStatus)) {
		debug("[-] NtUnmapViewOfSection error: %X\n", ntStatus);
		return 0;
	}

	// Close NTDLL section handle
	CloseHandle(hSection);

	// Close NTDLL disk handle
	CloseHandle(hFile);

	debug("[+] All done.\n");

	return 1;

}

// Technique - 2
// Reads NTDLL From KnownDlls and Clean
BOOL Technique2() {

	debug("[*] Using Technique-2, Reads NTDLL From KnownDlls and Clean.\n\n");

	NTSTATUS ntStatus;

	// Get handle to loaded/hooked NTDLL
	HMODULE hHookedNtdll = GetModuleHandleA("ntdll.dll");

	if (hHookedNtdll == NULL) {
		debug("[-] GetModuleHandleA error: %d\n", GetLastError());
		return FALSE;
	}

	OBJECT_ATTRIBUTES ObjectAttributes = { 0 };
	UNICODE_STRING ObjectPath = { 0 };
	IO_STATUS_BLOCK IoStatusBlock = { 0 };
	_RtlInitUnicodeString RtlInitUnicodeString = (_RtlInitUnicodeString)GetProcAddress(hHookedNtdll, "RtlInitUnicodeString");
	if (RtlInitUnicodeString == NULL) {
		return FALSE;
	}
	RtlInitUnicodeString(&ObjectPath, L"\\KnownDlls\\ntdll.dll");

	InitializeObjectAttributes(&ObjectAttributes, &ObjectPath, OBJ_CASE_INSENSITIVE, NULL, NULL);

	HANDLE hKnownDll = NULL;

	ntStatus = RxNtOpenSection(&hKnownDll, SECTION_MAP_READ | SECTION_MAP_EXECUTE, &ObjectAttributes);

	if (!NT_SUCCESS(ntStatus)) {
		debug("[-] NtOpenSection error.\n");
		CloseHandle(hKnownDll);
		return FALSE;
	}

	debug("[*] Clean Section Handle Address...........: 0x%p\n\n", hKnownDll);

	LPVOID pCleanNtdll = NULL;
	SIZE_T sztViewSize = 0;

	ntStatus = RxNtMapViewOfSection(hKnownDll, NtCurrentProcess(), &pCleanNtdll, (ULONG_PTR)NULL, (SIZE_T)NULL, NULL, &sztViewSize, 1, 0, PAGE_READONLY);

	if (!NT_SUCCESS(ntStatus)) {
		debug("[-] NtMapViewOfSection error.\n");
		CloseHandle(hKnownDll);
		return FALSE;
	}

	debug("[*] Clean NTDLL Base Address...............: 0x%p\n", pCleanNtdll);

	MODULEINFO miHookedNtdll = {};

	if (GetModuleInformation(NtCurrentProcess(), hHookedNtdll, &miHookedNtdll, sizeof(miHookedNtdll)) == 0) {
		debug("[-] GetModuleInformation error: %d\n", GetLastError());
		return FALSE;
	}

	// Get base address of hooked NTDLL from MODULEINFO struct
	LPVOID pHookedBaseAddress = (LPVOID)miHookedNtdll.lpBaseOfDll;
	debug("[*] Hooked NTDLL Base Address..............: 0x%p\n\n", pHookedBaseAddress);

	// Get hooked NTDLL DOS header
	PIMAGE_DOS_HEADER pHookedDosHeader = (PIMAGE_DOS_HEADER)pHookedBaseAddress;

	// Get hooked NTDLL NT header
	PIMAGE_NT_HEADERS pHookedNtHeader = (PIMAGE_NT_HEADERS)((DWORD_PTR)pHookedBaseAddress + pHookedDosHeader->e_lfanew);

	for (SIZE_T i = 0; i < pHookedNtHeader->FileHeader.NumberOfSections; i++) {

		// Get PE section header
		PIMAGE_SECTION_HEADER pHookedSectionHeader = (PIMAGE_SECTION_HEADER)((DWORD_PTR)IMAGE_FIRST_SECTION(pHookedNtHeader) + ((DWORD_PTR)IMAGE_SIZEOF_SECTION_HEADER * i));

		// Get section name
		LPSTR szHookedSectionName = (LPSTR)pHookedSectionHeader->Name;

		// We found .TEXT section
		if (!strcmp(szHookedSectionName, ".text")) {

			// Get start address of hooked .TEXT section
			LPVOID pHookedTextSectionAddress = (LPVOID)((DWORD_PTR)pHookedBaseAddress + (DWORD_PTR)pHookedSectionHeader->VirtualAddress);

			// Get start address of clean .TEXT section
			LPVOID pCleanTextStartAddress = (LPVOID)((DWORD_PTR)pCleanNtdll + (DWORD_PTR)pHookedSectionHeader->VirtualAddress);

			// Get size of .TEXT section
			SIZE_T sztTextSectionSize = pHookedSectionHeader->Misc.VirtualSize;

			debug("[*] Hooked NTDLL .TEXT Section VA..........: 0x%p\n", pHookedTextSectionAddress);
			debug("[*] Clean NTDLL .TEXT Section VA...........: 0x%p\n\n", pCleanTextStartAddress);
			debug("[*] Size of .TEXT Section..................: %zd\n", sztTextSectionSize);

			// Change original page protection of hooked NTDLL to RWX
			LPVOID lpBaseAddress = pHookedTextSectionAddress;
			SIZE_T uSize = sztTextSectionSize;

			_NtProtectVirtualMemory NtProtectVirtualMemory = (_NtProtectVirtualMemory)CustomGetProcAddress(pCleanNtdll, "NtProtectVirtualMemory");

			ULONG oldProtection;
			ntStatus = NtProtectVirtualMemory(NtCurrentProcess(), &lpBaseAddress, &uSize, PAGE_EXECUTE_READWRITE, &oldProtection);

			if (!NT_SUCCESS(ntStatus)) {
				debug("[-] NtProtectVirtualMemory - 1: Error.\n");
				return FALSE;
			}

			// Copy .TEXT section of clean NTDLL into hooked NTDLL .TEXT section
			memcpy(pHookedTextSectionAddress, pCleanTextStartAddress, sztTextSectionSize);

			// Revert back to original page protections of overwritten NTDLL .TEXT section
			ntStatus = NtProtectVirtualMemory(NtCurrentProcess(), &lpBaseAddress, &uSize, oldProtection, &oldProtection);
			if (!NT_SUCCESS(ntStatus)) {
				debug("[-] NtProtectVirtualMemory - 2: Error.\n");
				return FALSE;
			}

			debug("\n[+] NTDLL is cleaned. Closing handles...\n\n");

			break;

		}

	}

	// Unmap the local section
	_NtUnmapViewOfSection NtUnmapViewOfSection = (_NtUnmapViewOfSection)GetProcAddress(hHookedNtdll, "NtUnmapViewOfSection");
	ntStatus = NtUnmapViewOfSection(NtCurrentProcess(), pCleanNtdll);
	if (!NT_SUCCESS(ntStatus)) {
		debug("[-] NtUnmapViewOfSection error: %X\n", ntStatus);
		return FALSE;
	}

	// Close KnownDll handle
	CloseHandle(hKnownDll);

	debug("[+] All done.\n");

	return TRUE;

}
