# ds360go
xboxdrv wrapper to use a Dualsense as a Xbox 360 controller on Linux

# requirements

pgrep, xboxdrv, go, go-tools, make, systemd, udev

# building

make

# installation

sudo make install
sudo make reload

# running

ds360go should start when DualSense gets connected and stops when it is disconnected

this is accomplished with udev rules and systemd --user service. see source for more info.

# steam

ds360go will stop xboxdrv when it detects steam game overlay because steam has its own controller driver that interferes and older games get confused

# arch

paru -S ds360go-git

# license

MIT