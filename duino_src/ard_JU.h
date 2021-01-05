#ifndef _ARD_JU_ACH_
#define _ARD_JU_ACH_

typedef struct {
  char instruction;          // The instruction that came in over the serial connection.
  float argument;            // The argument that was supplied with the instruction.
} Command;

char serial_in;
int SMode = 0;
int wipeReverse = 0;

const byte nLEDS = 30;
const byte ledPin = 6;



#endif