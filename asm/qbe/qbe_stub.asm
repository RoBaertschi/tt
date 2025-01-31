; Reference: https://filippo.io/linux-syscall-table/
; https://blog.rchapman.org/posts/Linux_System_Call_Table_for_x86_64/
; https://en.wikipedia.org/wiki/X86_calling_conventions#List_of_x86_calling_conventions
format ELF64

section ".text" executable

    public syscall1
    public syscall2
    public syscall3
    ; rdi => Syscall number, rsi => argument
syscall1:
    mov rax, rdi
    mov rdi, rsi
    syscall
    ret

syscall2:
    mov rax, rdi
    mov rdi, rsi
    mov rsi, rdx
    syscall
    ret


syscall3:
    mov rax, rdi
    mov rdi, rsi
    mov rsi, rdx
    mov rdx, rcx
    syscall
    ret
