# ardpifi

# johnusher/ardpifi Summary

Mesh network with Raspis to control Arduino controlled LED strips using GPS and accelerometers.

A webserver on the pi provides a bi-directional UI (currently only via ethernet SSH to the headless pi).

LED code can be updated, compiled and flashed locally via USB from Raspi to Arduino using the duino CLI.

Keyboard inputs on each Raspi will send LED pattern info and sync across all Raspis in the network.




## Hardware set-up

See shopping list section for details on components.

### Arduino
Connect Arduino Uno (also Nano clone tested) to Raspi 3 via USB. Programmable NeoPixelLED strips connected to Arduino and should be powered separately:

### GPS
GPS module connect with GPIO serial (upin 15, 16, 5 V power to module). It shows up at serial port /dev/ttyS0. In raspi-config settings, you may need to disable serial console output, and disable bluetooth. I have used an external antenna and a small ceramic antenna: both seem to work.

Have tested the UBLOX NEO-6M (GPS only= less accurate) NEO M-9N (GPS, GLONASS, Galileo = more accurate).

NEO M-9N is default 38000 baud UART, but we change to 9600 and poll every 2 seconds. To this this, use u-center 21.02 (windows only), or use NEO9Settings.txt configuration script with ubxconfig.sh from https://gist.github.com/hdoverobinson/42732da4c5b50d031f60c1fae39f0720)  (untested!)

```bash
./ubxconfig.sh /dev/ttyS0 NEO9Settings.txt
```

Useful GPS:


Read raw NMEA format GPS data from serial port
```bash
cat -A /dev/ttyUSB0
```

Write raw NMEA format data to file
```bash
cat </dev/ttyUSB0 > gpsdata.txt
```

Convert NMEA file to Google-maps kml, discard low-HDOP data that is not "excellent quality", dude.
```bash
gpsbabel -i nmea -f gpsdata.txt -x discard,hdop=1,sat=9  -o kml -F outfile.kml
```


### Audio: I2S

install port audio and make I2s default:
```bash
sudo apt-get install libasound-dev portaudio19-dev libportaudio2 libportaudiocpp0 mpg123 
curl -sS https://raw.githubusercontent.com/adafruit/Raspberry-Pi-Installer-Scripts/master/i2samp.sh | bash
```
reboot.

Test audio works:

```bash
sudo speaker-test -c2 
sudo mpg123 -b 1024 http://ice1.somafm.com/u80s-128-mp3
```
to generate white noise out of the speaker, alternating left and right.

There is no built-in audio out with the Pi-Zero, so we use an I2S audio DAC: UDA1334.

from https://learn.adafruit.com/adafruit-i2s-stereo-decoder-uda1334a/raspberry-pi-wiring

Amp Vin to Raspbery Pi 3V or 5V

Amp GND to Raspbery Pi GND

Amp DIN to Raspbery Pi #21  = pin 21

Amp BCLK to Raspbery Pi GPIO#18 = pin 12

Amp LRCLK to Raspbery Pi #19 = pin 35


We must run audio playback as sudo! eg  go build wavestest.go && sudo ./wavestest 

Optional:

Edit /usr/share/alsa/alsa.conf to change pcm.front cards.pcm.front -> pcm.front cards.pcm.default

sudo apt-get --no-install-recommends install jackd2

### GPIO: Button and LED

An arcade push button attaches to GPIO27= physical pin 13, and an LED with 330 Ohm series resistor to GPIO22 = physcial pin 13.

GPIOTEST.go shows this in action.

go get github.com/warthog618/gpiod

GPIO support not intalled for Raspi Lite so run this:
```bash
sudo apt-get install wiringpi
gpio readall
```

### Display screen
OLED 128*64.

Power with +5 V. OLED connects via I2C bus 1. SDA -> pin 3  SCL -> pin 5). 

### Inertial Measurement Unit (IMU)
Bosch BNO055:  triaxial accelerometer, gyroscope, geomagnetic sensor.

Power with +3.3 V.
connect BNO055 SDA to pin 16 and SCLK to pin 18.

I2c configure:

From https://github.com/kpeu3i/bno055/:

"it seems all versions of Raspberry Pi have an I²C bus hardware problem preventing them from working correctly with Bosch BNO055. The problem has been variously diagnosed as being due to the Pi’s inability to handle clock stretching in arbitrary parts of the I²C transaction and the BNO055 chip’s exquisite sensitivity to I²C bus levels."


Raspbian has a software I2C driver that can be enabled by adding the following line to /boot/config.txt:

dtoverlay=i2c-gpio,bus=3
This will create an I2C bus called /dev/i2c-3. SDA will be on GPIO23 and SCL will be on GPIO24 which are pins 16 and 18 on the GPIO header respectively.

SDA = pin 16, SCL = pin 18.

```bash
sudo tee --append /boot/config.txt > /dev/null << 'EOF'
dtoverlay=i2c-gpio,bus=3
EOF
```
++sudo reboot now


Check your I2c is set up correctly:

sudo apt-get install i2c-tools

sudo i2cdetect -l

You will now see that i2c bus 3 also listed. Also run:

sudo i2cdetect -y 1

This should show OLED on 3c


sudo i2cdetect -y 3

This should show the BNo055 on 28.



There are "no hardware" options for running the main Go file without these HW modules.

Run the go script, and keyboard numbers will send a command to the mesh network, eg controlling LED sequence on the LED strip.

LED sequences can be programmed on the Raspi, compiled using arduino-cli, and flashed from the Raspi.

## WLAN dongle: 

Add USB dongle TP-Link TL-WN725N.

This messes up BATMAN as the dongle can go to WLAN0 or WLAN1.
To change this and make WLAN0 the built-in: see 

https://www.raspberrypi.org/forums/viewtopic.php?f=36&t=198946

1. switch off systemd-predictable mechanism stuff (unless done already):

```bash
sudo ln -nfs /dev/null /etc/systemd/network/99-default.link
  ```

  2. setup udev rules that uniquely identify your interfaces based on their USB connector positions:

```bash

vi /etc/udev/rules.d/72-wlan-geo-dependent.rules

#
# +---------------+
# | wlan1 | wlan1 |
# +-------+-------+
# | wlan1 | wlan1 |
# +---------------+ (RPI USB ports with position independent device names for a maximum of 1 optional wifi dongle)
# 
# | wlan0 | (onboard wifi)
#
ACTION=="add", SUBSYSTEM=="net", SUBSYSTEMS=="sdio", KERNELS=="mmc1:0001:1", NAME="wlan0"
ACTION=="add", SUBSYSTEM=="net", SUBSYSTEMS=="usb",  KERNELS=="1-1.2",       NAME="wlan1"
ACTION=="add", SUBSYSTEM=="net", SUBSYSTEMS=="usb",  KERNELS=="1-1.4",       NAME="wlan1"
ACTION=="add", SUBSYSTEM=="net", SUBSYSTEMS=="usb",  KERNELS=="1-1.3",       NAME="wlan1"
ACTION=="add", SUBSYSTEM=="net", SUBSYSTEMS=="usb",  KERNELS=="1-1.5",       NAME="wlan1"

 ```

NB

When running on WLAN1, we must force IP4 to access github.com as it does not support IP6

sudo dhclient -4 -v wlan1



or

```bash
sudo tee --append /etc/sysctl.conf > /dev/null << 'EOF'
net.ipv6.conf.all.disable_ipv6=1
EOF
 ```



echo sudo net.ipv6.conf.all.disable_ipv6=1 >> /etc/sysctl.conf


Install driver for tl-wn8725N

from:
https://www.raspberrypi.org/forums/viewtopic.php?p=462982#p462982
```bash
sudo wget http://downloads.fars-robotics.net/wifi-drivers/install-wifi -O /usr/bin/install-wifi
sudo chmod +x /usr/bin/install-wifi
sudo install-wifi
 ```
reboot


make sure you have added static ip to wlan 1:

```bash
sudo tee --append /etc/dhcpcd.conf > /dev/null << 'EOF'

# set static ip

interface wlan1
static ip_address=192.168.1.164/24
static routers=192.168.1.1
static domain_name_servers=192.168.1.1

EOF
 ```

after iwconfig
should see wlan0 and wlan connected to same ESSID with AP.


## Auto-start: 

Automatically load BATMAN ad the go script on boot.

To do this, we make a service. This runs on bootup, after the network is up. It runs runBATMAN.sh (or run runBATMAN2, runBATMAN3):

After 20 seconds, the BATMAN mesh network is generated, then the go app is run.

Thus, you need to make sure the app is pre-compiled,
ie
```bash
cd /home/pi/apl/
go build JU_led_mesh.go
```

There are many ways to do this, but this one is mine. systemctl is not my best friend. For about an hour, it was my life. I do not care to master it.

1. cd ~/apl2

2. chmod 664 delayBoot2.service

3. sudo cp   delayBoot2.service  /etc/systemd/system/

4. sudo systemctl enable  delayBoot2

5. sudo systemd-analyze verify delayBoot2.service

To reload systemd with this new service unit file, run:
systemctl daemon-reload


To start the script on boot, enable the service with systemd:

6. sudo systemctl enable delayBoot2.service

useful:

sudo systemctl stop delayBoot2.service

sudo systemctl start delayBoot2.service

systemctl status delayBoot2.service







## Hardware shopping list

Rasperry Pi 3 Model B+ tested.

Arduino clone:
https://www.christians-shop.de/Nano-V3-developer-board-for-Arduino-IDE-ATMEL-ATmega328P-AVR-Microcontroller-CH340-Chip-Christians-Technikshop

https://www.amazon.de/-/en/gp/product/B078SBBST6/ref=ppx_yo_dt_b_asin_title_o02_s00?ie=UTF8&psc=1

GPS: uBlox NEO-6M :
https://www.amazon.de/dp/B088LR3488/ref=pe_3044161_185740101_TE_item

GPS: uBlox NEO-M9N :
https://tinyurl.com/j2476wb  (Banggood.com)

OLED display:
https://www.amazon.de/-/en/gp/product/B01L9GC470/ref=ppx_od_dt_b_asin_title_s00?ie=UTF8&psc=1

IMU Bosch BNO055:
https://www.amazon.de/-/en/gp/product/B072NLTPTJ/ref=ppx_yo_dt_b_asin_title_o01_s00?ie=UTF8&psc=1

BROgrammable LED strip WS2812:
https://www.mouser.de/ProductDetail/digilent/122-000/?qs=WbxR7jUW5e9xhU9oZFzZgA==&countrycode=DE&currencycode=EUR

WLAN dongle TP-Link TL-WN725N:
https://www.amazon.de/gp/product/B008IFXQFU/ref=ppx_yo_dt_b_asin_title_o03_s00?ie=UTF8&psc=1




### OS

Install Raspi lite from onlione or try this:

1. Download Raspian Lite: https://downloads.raspberrypi.org/raspbian_lite_latest
2. Flash `20XX-XX-XX-raspbian-stretch-lite.zip` using Etcher
3. Remove/reinsert flash drive
4. Add `ssh` and `bootstrap` files:
    ```bash
    touch /Volumes/boot/ssh
    cp bin/bootstrap /Volumes/boot/
    chmod a+x /Volumes/boot/bootstrap
    diskutil umount /Volumes/boot
    ```

### First Boot

```bash
ssh pi@raspberrypi.local
# password: raspberry

# change default password
passwd


# set quiet boot
sudo sed -i '${s/$/ quiet loglevel=1/}' /boot/cmdline.txt

# install packages
sudo apt-get update
sudo apt-get install -y git tmux vim dnsmasq hostapd

# set up wifi (note leading space to avoid bash history)
sudo tee --append /etc/wpa_supplicant/wpa_supplicant.conf > /dev/null << 'EOF'
network={
    ssid="<WIFI_SSID>"
    psk="<WIFI_PASSWORD>"
}
EOF

# set static IP address
sudo tee --append /etc/dhcpcd.conf > /dev/null << 'EOF'

# set static ip

interface eth0
static ip_address=192.168.1.164/24
static routers=192.168.1.1
static domain_name_servers=192.168.1.1

interface wlan0
static ip_address=192.168.1.164/24
static routers=192.168.1.1
static domain_name_servers=192.168.1.1


EOF


if we have an additional wlan usb:

sudo tee --append /etc/dhcpcd.conf > /dev/null << 'EOF'

# set static ip

interface wlan1
static ip_address=192.168.1.164/24
static routers=192.168.1.1
static domain_name_servers=192.168.1.1


EOF


# reboot to connect over wifi
sudo shutdown -r now
```

```bash
# configure git
git config --global push.default simple
git config --global core.editor "vim"
git config --global user.email "you@example.com"
git config --global user.name "Your Name"

# disable services
sudo systemctl disable hciuart
sudo systemctl disable bluetooth
sudo systemctl disable plymouth

# remove unnecessary packages
sudo apt-get -y purge libx11-6 libgtk-3-common xkb-data lxde-icon-theme raspberrypi-artwork penguinspuzzle ntp plymouth*
sudo apt-get -y autoremove

sudo raspi-config nonint do_boot_behaviour B2 0
sudo raspi-config nonint do_boot_wait 1
sudo raspi-config nonint do_serial 1
```

# now you are setup! run bootstrap to download packages

Run Bootstrap from https://github.com/johnusher/ardpifi/tree/master/bin
./bootstrap
```




### Mesh network

NB need to runt his command twice!

On raspi #1 run

./runBATMAN.sh

On raspi #2 run

./runBATMAN2.sh



## Code

Note
All dependencies managed in `go.mod` now,
just add an import directive for any new depedency in your `*.go` files, and
`go run/build` should just handle it.


All code (and go) is installed via bootstrap.



# Arduino CLI install

follow instructions here:
https://siytek.com/arduino-cli-raspberry-pi/

this additional command was needed:
arduino-cli core install arduino:avr

Note the directory for the Duino project must have the same name as the main file .ino.

Tested with Duino Uno and Duino clone: "Nano V3 | ATMEL ATmega328P AVR Microcontroller | CH340-Chip".
The Uno shows on port ttyACM0 and the clone on ttyUSB.


## add libraries:

arduino-cli lib search Adafruit_NeoPixel


in duino_src:

compile and flash:
Uno:
arduino-cli compile --fqbn arduino:avr:uno duino_src

arduino-cli upload -p /dev/ttyACM0 --fqbn arduino:avr:uno duino_src

Clone

arduino-cli compile --fqbn arduino:avr:diecimila:cpu=atmega328 duino_src

arduino-cli upload -p /dev/ttyUSB0 --fqbn arduino:avr:diecimila:cpu=atmega328 duino_src

arduino-cli upload -p /dev/ttyACM0 --fqbn arduino:avr:diecimila:cpu=atmega328 duino_src


## Run


On raspi #1 run

```bash
./runBATMAN.sh
go run JU_led_mesh.go -rasp-id=me --web-addr :8080 -log-level debug
```

On raspi #2 run

```bash
./runBATMAN2.sh
go run JU_led_mesh.go -rasp-id=poo --web-addr :8081 -no-lcd -log-level debug.
```

Press any key, sent to mesh, and if it is a 0 or 1, we change led pattern.

To exit, press "q" to exit termbox, and then ctrl-c to exit the program.

There are various flags if certain HW modules are not attached to the raspi:


### Flags

```bash
$ go run JU_led_mesh.go -h
  -log-level string
    	log level, must be one of: panic, fatal, error, warn, info, debug, trace (default "info")
  -no-batman
    	run without batman network
  -no-duino
    	run without arduino
  -no-gps
    	run without gps
  -no-acc
    	run without Bosch accelerometer      
  -no-oled
    	run without oled display      
  -rasp-id string
    	unique raspberry pi ID (default "raspi 1")
  -web-addr string
    	address to serve web on (default ":8080")
```

### Run without hardware

Run with hardware (serial, network) API calls mocked out:

```bash
go run JU_led_mesh.go --rasp-id "raspi 1" --web-addr :8080 --no-batman --no-duino --no-gps --no-lcd --log-level debug
```

# Set up port forwarding for web server

## First-time ssh config setup

On laptop:


```bash
PI_IP=192.168.1.164
USER=pi

cat << EOF >> ~/.ssh/config

Host pi
  HostName $PI_IP
  User $USER
  Port 22
  BatchMode yes
  ServerAliveInterval 60
  ServerAliveCountMax 5
  ForwardAgent yes
    LocalForward 127.0.0.1:8080 127.0.0.1:8080
EOF
```

## Tunnel

```bash
ssh $USER@pi
```

Open a browser window to localhost:8080


## Other useful scripts
```bash
run BNo055_save_data.go
```
This saves accelerometer and gyroscope data to a txt file.

```bash

go build GPIOTEST.go && sudo ./GPIOTEST

```
This reads the button, debounces, and plays sound when we have button down.


```bash
run OLEDtest.go
```
This displays a real-time clock on the OLED.

convAccData.m

MATLAB script to process accelerometer data.


# thanks 

@siggy!