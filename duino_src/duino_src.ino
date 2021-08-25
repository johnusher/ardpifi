/**
  Arduino code:
  receive message on serial USB and change LEDs

*/


#include "ard_JU.h"

#include "Adafruit_NeoPixel.h"   // NB this should be from https://github.com/adafruit/Adafruit_NeoPixel!

Adafruit_NeoPixel strip = Adafruit_NeoPixel(nLEDS, ledPin, NEO_GRB + NEO_KHZ800);

void setup() {
  pinMode(ledPin, OUTPUT);
  strip.begin();
  myStripShow(); // Initialize all pixels to 'off'
  // Serial.begin(9600);
  // Serial.begin(19200);
  Serial.begin(115200);

  idleCol =  strip.Color(idleColR, idleColG, idleColB);
}

/**
   Main Arduino loop.
*/
void loop() {

  // Serial.print("S");

  // for (i = 0; i < strip.numPixels(); i++) {
  //   strip.setPixelColor(i, idleColR, idleColG, idleColB, 0);
  //   //    strip.setPixelColor(i, 0, 0, 0, 127);
  //   myStripShow();
  //   // Serial.print("x");
  //   //        tone(buzzer, i + j); // Send xm sound signal...

  // }

  idleC = 0;
  while (Serial.available() == 0) {

    // waiting for serial command
    idleC = idleC + 1;
    idleC = idleC % strip.numPixels();

    // if(idleC==strip.numPixels()){
    // colC = colC + 1;
    // colC = colC % 127;

    // // change color:
    //   uint8_t idleColR = 0; 	// colour of idle
    //   uint8_t idleColG = 64%colC;
    //   uint8_t idleColB = 127%colC;

    // }

    strip.setPixelColor(idleC, idleColR, idleColG, idleColB);  // show blue light moving along

    if (breakFlag == 0) {

      myStripShow();

      for (i = 0; i < 1000; i++) {
        checkSerialInput();

      }

    }

    else
    {
      idleC = 0;

    }


    breakFlag = 0;

    // strip.setPixelColor(idleC, idleColR, idleColG, idleColB);
    // myStripShow();

    //
    //    // rainbow:
    //    for (j = 0; j < maxC; j++) {
    //      for (i = 0; i < strip.numPixels(); i++) {
    //        strip.setPixelColor(i, Wheel((i + j) & (maxC-1)));
    //        // Serial.print("x");
    //        //        tone(buzzer, i + j); // Send xm sound signal...
    //
    //      }
    //      //      myStripShow();
    //      myStripShow();
    ////      delay(5);
    //    }



  }

  checkSerialInput();





  //  if (Serial.available() > 0) {
  //    // read the incoming byte:
  //    char c = Serial.read();
  //
  //    // say what you got:
  //    Serial.print("I received: ");
  //    Serial.println(c);
  //  }
}



void checkSerialInput() {
  idle_flag = 0;
  strip.clear();
  serial_in = Serial.read();

  // say what you got:
  //  Serial.print(serial_in);
  //  Serial.flush();

  if (serial_in == '0') {
    SMode = 0;
    Serial.println("mode 0");
    Serial.flush();
    Serial.println("a");

    idleC = 128;
    idleColR = 0;
    idleColG = 0;
    idleColB = 128;


    strip.setPixelColor(idleC, idleColR, idleColG, idleColB);  // show blue light moving along
    Serial.println("b");


    
//    if (breakFlag == 1) {
//      //    colorWipe(strip.Color(0, 0 , 0), 50); // off
//      Serial.println("c");
//    }
//    else {
//      wipeReverse = !wipeReverse;
//      Serial.println("d");
//    }
//    Serial.println("e");
    breakFlag = 0;
    idleC = 0;

  }

  else if (serial_in == '1') {
    SMode = 1;
    Serial.println("mode 1");
    Serial.flush();

    colorWipe(strip.Color(0, 0 , maxC ), 25); // g
    Serial.print("breakFlag = ");
    Serial.println(breakFlag);


    if (breakFlag == 1) {
      colorWipe(strip.Color(0, 0 , 0), 50); // off
    }
    else {
      wipeReverse = !wipeReverse;
    }
    breakFlag = 0;
    idleC = 0;

  }



}

void myStripShow() {
  if (Serial.available() > 0) {
    checkSerialInput();
    breakFlag = 1;
  }
  strip.show();
}


// Fill the dots one after the other with a color
void colorWipe(uint32_t c, uint8_t wait) {

  breakFlag = 0;
  if (wipeReverse)
  {
    for (uint16_t i = 0; i < strip.numPixels(); i++) {
      strip.setPixelColor(i, c);
      //      tone(buzzer, i ); // Send 1KHz sound signal...
      myStripShow();
      if (breakFlag == 1) {
        break;
      }
      else {
        delay(wait);
      }
    }
  }
  else
  {
    for (uint16_t i = 0; i < strip.numPixels(); i++) {
      strip.setPixelColor(strip.numPixels() - i, c);
      //      tone(buzzer, i XI); // Send 1KHz sound signal...

      myStripShow();
      if (breakFlag == 1) {
        break;
      }
      else {
        delay(wait);
      }
    }
  }

  wipeReverse = !wipeReverse;


  //  noTone(buzzer);     // Stop sound...


}


// Slightly different, this makes the rainbow equally distributed throughout
void rainbowCycle(uint8_t wait) {
  uint16_t i, j;

  for (j = 0; j < 256 * 5; j++) { // 5 cycles of all colors on wheel
    for (i = 0; i < strip.numPixels(); i++) {
      strip.setPixelColor(i, Wheel(((i * 256 / strip.numPixels()) + j) & 255));

      //     tone(buzzer, i + j); // Send xm sound signal...

    }
    myStripShow();
    delay(wait);
  }

}

// Input a value 0 to 255 to get a color value.
// The colours are a transition r - g - b - back to r.
uint32_t Wheel(byte WheelPos) {
  if (WheelPos < 85) {
    return strip.Color(WheelPos * 3, 255 - WheelPos * 3, 0);
  } else if (WheelPos < 170) {
    WheelPos -= 85;
    return strip.Color(255 - WheelPos * 3, 0, WheelPos * 3);
  } else {
    WheelPos -= 170;
    return strip.Color(0, WheelPos * 3, 255 - WheelPos * 3);
  }
}
