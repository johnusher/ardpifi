# ardpifi # ardpifi # ardpifi


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
(install go)


mkdir -p ~/code/go/src/github.com/johnusher
git clone https://github.com/johnusher/ardpifi.git ~/code/go/src/github.com/johnusher/ardpifi