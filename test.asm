format ELF64 executable

segment readable executable
main:
    mov rax, 0
    ret

entry _start
_start:
    call main
    mov rdi, rax
    mov rax, 60
    syscall
