# T-Display S3 with tinygo
This repository contains code and documentation for using the T-Display S3 with tinygo.

It contains the following packages:
* [`display`](https://pkg.go.dev/github.com/axellse/tdisplays3-tinygo/display) - easy to use display driver implementing `drivers.Displayer`.
* [`i8080`](https://pkg.go.dev/github.com/axellse/tdisplays3-tinygo/i8080) - generic i8080 interface implementing `drivers.SPI`.

## Display
The T-Display S3 has a beautiful ST7789-driven 170 by 320 RGB display. You talk to it with the Intel 8080 interface (aka. "I8080", "I80" and "8 bit parallel").

### ST7789 driver
Tinygo's amazing driver repo comes in clutch again and has a [built in driver for this](https://pkg.go.dev/tinygo.org/x/drivers/st7789). It can be configured like so:
```go
//GPIOs and stuff is in aliexpress listing which i wont bother linking to because aliexpress links never work
display := st7789.New(iface, machine.GPIO5, machine.GPIO7, machine.GPIO6, machine.GPIO38)
display.Configure(st7789.Config{
	ColumnOffset: 35, //(240 - 170) / 2 = 35, thank you https://homeding.github.io/boards/esp32s3/lilygo-t-display-s3.htm
	Width:        170,
	Height:       320,
	Rotation:     drivers.Rotation90, //The display is in portrait by default, this puts it in landscape. For me, a rotation of zero does not work properly for some reason.
})
```
### I8080 interface
The ESP32-S3 used by the T-Display S3 has a built-in hardware component for addressing I8080 displays (called LCD_CAM), but i have not found nor written a tinygo driver for this yet. If you're not doing animation, playing video or need fast display updates, a simple software I8080 implementation along with a framebuffer on the esp32 works fine. This setup can achieve full screen updates in about 1/3 of a second.

The ST7789 driver is built to work with the ST7789 over SPI (the chip support both SPI and I8080) and thus wants something implementing drivers.SPI. I have written a very simple "bit-banged" I8080 implementation [in here](https://pkg.go.dev/github.com/axellse/tdisplays3-tinygo/i8080). It can be configured like so:
```go
iface := i8080.SoftSPI{
	WRPin:    machine.GPIO8,
	RDPin:    machine.GPIO9,
	DataPins: []machine.Pin{machine.GPIO39, machine.GPIO40, machine.GPIO41, machine.GPIO42, machine.GPIO45, machine.GPIO46, machine.GPIO47, machine.GPIO48},
}

err := iface.Init()
if err != nil {
	panic(err)
}
```
### Framebuffer
Unlike many Tinygo display drivers, the ST7789 driver does not include a frame buffer because of the memory footprint. For the T-Display S3, that works out to about 108KB (170 * 320 * 2). As of writing this tinygo can't yet utilize the 8MB of PSRAM on the T-Display S3, **this means about a third of the usable internal memory is being used for this buffer**. [There is an ongoing effort to add PSRAM support though](https://github.com/tinygo-org/tinygo/pull/5554).

It is possible to use the display without a buffer, [but it makes a big difference](https://www.youtube.com/watch?v=Xt91Vq3EHAE).

You can use `pixel.Image[pixel.RGB565BE]` for the framebuffer.