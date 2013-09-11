# Skyline's Algorithm

golang version based on https://github.com/etsy/skyline/blob/master/src/analyzer/algorithms.py#L230

# Install cephes

1. download cephes

    wget http://www.netlib.org/cephes/{cmath.tgz,eval.tgz, cprob.tgz}

1. edit cprob.mak

    # Makefile for probability integrals.
    # Be sure to set the type of computer and endianness in mconf.h.

    CC = gcc
    CFLAGS = -fPIC -O2 -Wall
    INCS = mconf.h

    OBJS = bdtr.o btdtr.o chdtr.o drand.o expx2.o fdtr.o gamma.o gdtr.o \
igam.o igami.o incbet.o incbi.o mtherr.o nbdtr.o ndtr.o ndtri.o pdtr.o \
stdtr.o unity.o polevl.o const.o

    libprob.so: $(OBJS) $(INCS)
        gcc -shared  $(OBJS) -o libprob.so

1. compile libprob.so

    make -f cprobe.mak

# LICENSE

MIT LICENSE
