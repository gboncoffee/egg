branches_test:
add t0, t1, t2
addi t0, t1, 2
addiu t0, t1, 3
addu t0, t1, t2
clo t0, t1
clz t0, t1
lui t0, 0xffff
seb t0, t1
seh t0, t1
sub t0, t1, t2
subu t0, t1, t2
sll t0, t1, 2
sllv t0, t1, t2
sra t0, t1, 2
srav t0, t1, t2
srl t0, t1, 2
srlv t0, t1, t2
and t0, t1, t2
andi t0, t1, 0xff
nor t0, t1, t2
or t0, t1, t2
ori t0, t1, 0xff
xor t0, t1, t2
xori t0, t1, 0xff
movn t0, t1, t2
movz t0, t1, t2
slt t0, t1, t2
slti t0, t1, 0xff
sltiu t0, t1, 0xff
sltu t0, t1, t2
div t0, t1
mult t0, t1
mfhi t0
mflo t0
mthi t0
mtlo t0
beq t0, t1, branches_test
bgez t0, branches_test
bgtz t0, branches_test
blez t0, branches_test
bltz t0, branches_test
bne t0, t1, branches_test
break
syscall
j 8
jal 8
jalr t0
jr t0
lb t0, t1, 0
lbu t0, t1, 0
lh t0, t1, 0
lhu t0, t1, 0
lw t0, t1, 0
sb t0, t1, 0
sh t0, t1, 0
sw t0, t1, 0
