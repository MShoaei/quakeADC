# GNU/Linux alongside Android

Device MUST be rooted. (e.g. using Odin and CF-Auto-Root)

install **[linux deploy](https://play.google.com/store/apps/details?id=ru.meefik.linuxdeploy&hl=en)** and busy box [this (meefik)](https://play.google.com/store/apps/details?id=ru.meefik.busybox&hl=en) or [this (stephen)](https://play.google.com/store/apps/details?id=stericson.busybox&hl=en).

to partition the SD card using the device itself use **[AParted](https://play.google.com/store/apps/details?id=com.sylkat.AParted&hl=en)**

also for ssh connection from the device to itself use **[JuiceSSH](https://play.google.com/store/apps/details?id=com.sonelli.juicessh&hl=en)**

---

There are 2 main ways of installing linux:

1. **(Recommended)** Create a custom image and copy the image to the SD Card. Then inside *Linux Deploy* set the *installation path* to point to the image and start the container. Also you may want to enable VNC and SSH for remote access to the device. (Couldn't get the *VNC* to work properly!)

2. **(Needs some work)** This one works but the process fails because _`/bin/su`_ is not found.
