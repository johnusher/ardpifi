# ardpifi 

# johnusher/ardpifi

Task 1: For different keystrokes, send associated I2C codeload.


# thanks @siggy!

### First Boot

```bash
ssh pi@raspberrypi.local
# password: x

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
static ip_address=192.168.1.141/24
static routers=192.168.1.1
static domain_name_servers=192.168.1.1

interface wlan0
static ip_address=192.168.1.142/24
static routers=192.168.1.1
static domain_name_servers=192.168.1.1
EOF

# reboot to connect over wifi
sudo shutdown -r now


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



## Code

```bash

go get -u github.com/d2r2/go-logger

go get -u github.com/nsf/termbox-go

url='https://golang.org'$(curl https://golang.org/dl/ | grep armv6l | sort --version-sort | tail -1 | grep -o -E "/dl/go[0-9]+\.[0-9]+((\.[0-9]+)?).linux-armv6l.tar.gz")
archive_name=$(echo ${url} | cut -d '/' -f5)
wget ${url}
sudo tar -C /usr/local -xvf- ${archive_name}
cat >> ~/.bashrc << 'EOF'
export GOPATH=$HOME/go
export PATH=/usr/local/go/bin:$PATH:$GOPATH/bin
EOF
rm ${archive_name}
source ~/.bashrc


mkdir -p ~/code/go/src/github.com/johnusher
git clone https://github.com/johnusher/ardpifi.git ~/code/go/src/github.com/

git clone https://github.com/d2r2/go-i2c.git ~/go/src/github.com/d2r2/go-i2c

git clone https://github.com/nsf/termbox-go.git ~/go/src/github.com/nsf/termbox-go

export GOPATH=$HOME/code/go/src

cat >> ~/.bashrc << 'EOF'
export GOPATH=$HOME/code/go/src
export PATH=/usr/local/go/bin:$PATH:$GOPATH/bin
EOF

source ~/.bashrc