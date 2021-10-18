# ardpifi

# johnusher/ardpifi Summary

Mesh network with Raspis to control Arduino controlled LED strips using GPS and accelerometers.

A webserver on the pi provides a bi-directional UI (currently only via ethernet SSH to the headless pi).

LED code can be updated, compiled and flashed locally via USB from Raspi to Arduino using the duino CLI.

Keyboard inputs on each Raspi will send LED pattern info and sync across all Raspis in the network.




## Hardware set-up

![HardWare](https://github.com/johnusher/ardpifi/blob/master/ardPiFi_hw.png?raw=true)
  

See [shopping list section](#hardware-shopping-list) for details on components. 

### MCU for programmable LEDs 
ATmega328p- Arduino Nano. MCU and NeoPixelLED strips should be powered separately.

### GPS
GPS module connect with GPIO serial. It shows up at serial port /dev/ttyS0. In raspi-config settings, you  need to disable serial console output, enable Hardware serial, and maybe disable bluetooth. 

Tested using the UBLOX NEO-6M (GPS only= less accurate) and NEO M-9N (GPS, GLONASS, Galileo = more accurate).

Note we only use serial Rx from GPS module into the Raspi, as we use Serial Tx for Arduino MCU.

NEO M-9N default 38000 baud UART, but changed to 9600 and poll every 2 seconds. To do this this, use u-center 21.02 (windows only), or use NEO9Settings.txt configuration script with ubxconfig.sh from https://gist.github.com/hdoverobinson/42732da4c5b50d031f60c1fae39f0720)  (untested and won't work without serial Tx from Raspi)

```bash
./ubxconfig.sh /dev/ttyS0 NEO9Settings.txt
```

Useful GPS:
```bash
go run test_GPS2.go 
```

Prints Lat-Long cooridinates and HDOP.<br />
where HDOP:<br />
<1	Ideal	Highest possible confidence level to be used for applications demanding the highest possible precision at all times.<br />
1-2	Excellent	At this confidence level, positional measurements are considered accurate enough to meet all but the most sensitive applications.<br />
2-5	Good	Represents a level that marks the minimum appropriate for making accurate decisions. Positional measurements could be used to make reliable in-route navigation suggestions to the user.<br />
5-10	Moderate	Positional measurements could be used for calculations, but the fix quality could still be improved. A more open view of the sky is recommended.<br />


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

Raspi Zero does not have audio IO, so output is I2S into a 2 W amp, and then to a loudspeaker.

Install port audio and make I2s default:
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


We must run audio playback as sudo! eg  go build wavestest.go && sudo ./wavestest 

Optional:

Edit /usr/share/alsa/alsa.conf to change pcm.front cards.pcm.front -> pcm.front cards.pcm.default

sudo apt-get --no-install-recommends install jackd2

### GPIO: Button and LED

An arcade push button attaches to GPIO27=  pin 13, and an LED with 330 Ohm series resistor to GPIO22 =  pin 15.


GPIO support not intalled for Raspi Lite so run this:
```bash
sudo apt-get install wiringpi
gpio readall
```

on pi 4, try the following:

```bash
cd /tmp
wget https://project-downloads.drogon.net/wiringpi-latest.deb
sudo dpkg -i wiringpi-latest.deb
```

test button LED works:
```bash
gpio -g mode 22 out
gpio -g write 22 1
 ``` 
 
run this command to set pullup:
```bash
raspi-gpio set 27 pu
```

when you run
gpio readall
you should see following when button not pressed 
```bash
BCM| wPi |   Name  | Mode | V
27 |   2 | GPIO. 2 |   IN | 1 
```
```bash
27 |   2 | GPIO. 2 |   IN | 0
```
when pressed 



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

Rasperry Pi 3 Model B+ tested, Pi Zero WH, and Pi-4 B

MCU:
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

## Main routines to run:

```bash
go build test_record_spell.go && sudo ./test_record_spell -no-sound
```
test_record_spell records MCU movement when we have a button press.

NB The GPIO channel can not close properly, so you may need to kill the app with

sudo ps aux | grep test_record_spell

kill -9 -1 3290 

## Other useful scripts
```bash
go run test_BNo055_save_data.go
```
NB not running!! This saves accelerometer and gyroscope data to a txt file.

```bash

go build test_GPIO.go && sudo ./test_GPIO

```
NV this currently fails! This reads the button, debounces, and plays sound when we have button down.


```bash
go run OLEDtest.go
```
This displays a real-time clock on the OLED.

```bash
go run test_quats_64_to_py_tf.go
```
process quat recordings with TF... this tests if you have python and go correctly installed.


```bash
go run test_parse_quats_oled.go
```
test_parse_quats_oled





convAccData.m

MATLAB script to process accelerometer data.


# thanks 

@siggy!