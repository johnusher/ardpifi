/**


*/


#include "ard_JU.h"

#include <Adafruit_NeoPixel.h>

Adafruit_NeoPixel strip = Adafruit_NeoPixel(nLEDS, ledPin, NEO_GRB + NEO_KHZ800);

void setup() {
  pinMode(ledPin, OUTPUT);
  strip.begin();
  strip.show(); // Initialize all pixels to 'off'
  Serial.begin(9600);
}

/**
   Main Arduino loop.
*/
void loop() {

  while (Serial.available() == 0) {
    // waiting for serial command
  colorWipe(strip.Color(0, 255 , 0), 50); // g
  }

  serial_in = Serial.read();

  // say what you got:
  Serial.print("I received: ");
  Serial.println(serial_in);

  if (serial_in == '0') {
    SMode = 0;
    colorWipe(strip.Color(255, 0 , 0), 50); // Red
    Serial.println("mode 0");
  }

  if (serial_in == '1') {
    SMode = 1;
    colorWipe(strip.Color(0, 0 , 255), 50); // b
    Serial.println("mode 1");
  }


  //  if (Serial.available() > 0) {
  //    // read the incoming byte:
  //    char c = Serial.read();
  //
  //    // say what you got:
  //    Serial.print("I received: ");
  //    Serial.println(c);
  //  }
}





// Fill the dots one after the other with a color
void colorWipe(uint32_t c, uint8_t wait) {

  if (wipeReverse)
  {
    for (uint16_t i = 0; i < strip.numPixels(); i++) {
      strip.setPixelColor(i, c);
//      tone(buzzer, i ); // Send 1KHz sound signal...
      strip.show();
      delay(wait);
    }
  }
  else
  {
    for (uint16_t i = 0; i < strip.numPixels(); i++) {
      strip.setPixelColor(strip.numPixels() - i, c);
//      tone(buzzer, i XI); // Send 1KHz sound signal...

      strip.show();
      delay(wait);
    }
  }

  wipeReverse = !wipeReverse;


//  noTone(buzzer);     // Stop sound...


}
