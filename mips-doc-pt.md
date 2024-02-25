[Read this document in English](mips-doc.md)

# MIPS32 para o EGG

A implementação de MIPS32 para o EGG usa a sintaxe de Assembly padrão descrita
no [README](README-pt.md).

Para realizar uma chamada (_environment call_), o número dela é colocado no
registrador `v0` e os argumentos nos registradores `a0` e `a1`.

Uma instrução `break` realiza a chamada BREAK.

O programa montado é carregado no endereço `0` e o `pc` é inicializado em `0`.
