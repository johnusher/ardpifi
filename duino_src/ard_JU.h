#ifndef _ARD_JU_ACH_
#define _ARD_JU_ACH_

typedef struct {
  char instruction;          // The instruction that came in over the serial connection.
  float argument;            // The argument that was supplied with the instruction.
} Command;

char serial_in;
int SMode = 0;
int wipeReverse = 0;

int idleC = 0;	// counter for led position
int colC = 0;	// counter for led colour

uint32_t idleCol;
uint8_t idleColR = 255; 	// colour of idle
uint8_t idleColG = 112; 
uint8_t idleColB = 197; 


uint8_t maxC = 128;   // max colour value

bool idle_flag = 1; 

const byte nLEDS = 30;
const byte ledPin = 6;

int breakFlag = 0; 

uint16_t i, j;	// for rainbow effect


#endif
