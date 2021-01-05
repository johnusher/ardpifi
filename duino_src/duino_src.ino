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
    strip.show(); // Initialize all pixels to 'off'


  }

  serial_in = Serial.read();

  // say what you got:
  Serial.print("I received: ");
  Serial.println(serial_in);

  if (serial_in == '0') {
    SMode = 0;
    Serial.println("mode 0");
    colorWipe(strip.Color(255, 0 , 0), 25); // Red
    colorWipe(strip.Color(0, 0 , 0), 50); // off
    
  }

  if (serial_in == '1') {
    SMode = 1;
    Serial.println("mode 1");
    colorWipe(strip.Color(0, 0 , 255), 25); // b
    colorWipe(strip.Color(0, 0 , 0), 50); // off
    
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



// Slightly different, this makes the rainbow equally distributed throughout
void rainbowCycle(uint8_t wait) {
  uint16_t i, j;

  for (j = 0; j < 256 * 5; j++) { // 5 cycles of all colors on wheel
    for (i = 0; i < strip.numPixels(); i++) {
      strip.setPixelColor(i, Wheel(((i * 256 / strip.numPixels()) + j) & 255));

      //     tone(buzzer, i + j); // Send xm sound signal...

    }
    strip.show();
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
