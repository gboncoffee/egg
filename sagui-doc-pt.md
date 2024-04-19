[Read this document in English](sagui-doc.md)

# Sagui para o EGG

Sagui é uma arquitetura fantasia de 8 bits RISC, criada pelo Dr. Marco Zanata,
professor da Universidade Federal do Paraná. A implementação no EGG possui uma
extensão: a instrução `movr r0, r0` é interpretada como uma chamada BREAK.

Ela usa a sintaxe padrão de Assembly descrita no [README](README-pt.md). O
assembler é bem burro, infelizmente. Ele não corrige nenhum imediato de endereço
relativo, então basicamente você está por conta para computar todos os saltos
manualmente. Boa sorte.

![Instruções do Sagui](sagui/sagui.png)
