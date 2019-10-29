# Raspberry pi and ADC7768-4

## Implementation details

### SPI Interface

Each SPI access frame is **16 bits** long. The MSB (Bit **15**) of the
SDI command is the R/W bit; 1 = read and 0 = write. Bits[14:8] (**7** bits)
of the SDI command are the address bits.

### Data Interface

The data interface format is determined by setting the FORMATx
pins. The logic state of the FORMATx pins are read on power-
up and determine how many data lines (DOUTx) the ADC
conversions are output on.

Each ADC result comprises
32 bits. The first eight bits are the header status bits, which
contain status information and the channel number. The names
of each of the header status bits are shown in Table 35, and their
functions are explained in the subsequent sections. This header
is followed by a 24-bit ADC output in twos complement coding,
MSB first.

---

### Channel Modes

Using the channel mode select register (Register 0x03), the user can assign each channel to either Channel Mode A or Channel Mode B, which maps that mode to the required ADC channels.

### Interface Configuration

On the AD7768-4, it is recommended that Channel Mode A be set to the sinc5 filter whenever possible.

The DOUTx configuration for the AD7768-4 is selected using the FORMAT0 pin (see Table 34).

### CRC Protection

The AD7768/AD7768-4 can be configured to output a CRC message per channel every 4 or 16 samples. This function is available only with SPI control. CRC is enabled in the interface control register, Register 0x07 (see the CRC Check on Data Interface section).

### ADC Synchronization over SPI

To initiate the synchronization in this manner, write to Bit 7 in Register 0x06 twice.
First, the user must write a 0, which sets SYNC_OUT low, and then write a 1 to set the SYNC_OUT logic high again. The SPI_SYNC command is recognized after the last rising edge of SCLK in the SPI instruction, where the SPI_SYNC bit is changed from low to high. The SPI_SYNC command is then output synchronously to the AD7768/AD7768-4 MCLK signal on the SYNC_OUT pin. The user must connect the SYNC_OUT signal to the SYNC_IN pin on the PCB. Any daisy-chained system of AD7768/AD7768-4 devices requires that all ADCs be synchronized.

| FORMAT0 |                                  Description                                  |
| :-----: | :---------------------------------------------------------------------------: |
|    0    | Each ADC channel outputs on its own dedicated pin. DOUT0 to DOUT3 are in use. |
|    1    |  All channels output on the DOUT0 pin, in TDM output. Only DOUT0 is in use.   |

Each ADC channel outputs on its own dedicated
pin. DOUT0 to DOUT3 are in use.
All channels output on the DOUT0 pin, in TDM
output. Only DOUT0 is in use.

### ADC CONVERSION OUTPUT: HEADER AND DATA

The AD7768-4 data is output on the DOUT0 to DOUT3 pins, depending on the FORMAT0 pin.
The actual structure of the data output for each ADC result is shown in Figure 99.

Each ADC result comprises **32 bits**. The first **8** bits are the header status bits, which contain status information and the channel number. The names of each of the header status bits are shown in Table 35, and their functions are explained in the subsequent sections. This header is followed by a **24-bit** ADC output in twos complement coding, MSB first.

#### Table 35. Header Status Bits

| Bit |      Bit Name      |
| :-: | :----------------: |
|  7  |   ERROR_FLAGGED    |
|  6  | Filter not settled |
|  5  |   Repeated data    |
|  4  |    Filter type     |
|  3  |  Filter saturated  |

[2:0] | Channel ID[2:0]

#### Table 36. Channel ID vs. Channel Number

|  Channel  | Channel ID 2 | Channel ID 1 | Channel ID 0 |
| :-------: | :----------: | :----------: | :----------: |
| Channel 0 |      0       |      0       |      0       |
| Channel 1 |      0       |      0       |      1       |
| Channel 2 |      0       |      1       |      0       |
| Channel 3 |      0       |      1       |      1       |
| Channel 4 |      1       |      0       |      0       |
| Channel 5 |      1       |      0       |      1       |
| Channel 6 |      1       |      1       |      0       |
| Channel 7 |      1       |      1       |      1       |

## Commands

---

### Data Control

#### SPI_SYNC

Software synchronization of the AD7768-4. This command has the same
effect as sending a signal pulse to the START pin. To operate the SPI_SYNC,
the user must write to this bit two separate times. First, write a zero,
putting SPI_SYNC low, and then write a 1 to set SPI_SYNC logic high
again. The SPI_SYNC command is recognized after the last rising edge of
SCLK in the SPI instruction where the SPI_SYNC bit is changed from low to
high. The SPI_SYNC command is then output synchronous to the AD7768-4
MCLK on the SYNC_OUT pin. The user must connect the SYNC_OUT signal
to the SYNC_IN pin on the PCB. The SYNC_OUT pin can also be routed to
the SYNC_IN pins of other AD7768-4 devices, allowing larger channel
count simultaneous sampling systems. As per any synchronization pulse
seen by the SYNC_IN pin, the digital filters of the AD7768-4 are reset. The
full settling time of the filters must elapse before data is output on the
data interface. In a daisy-chained system of AD7768-4 devices, two
successive synchronization pulses must be applied to guarantee that all
ADCs are synchronized. Two synchronization pulses are also required in a
system of more than one AD7768-4 device sharing a single MCLK signal,
where the DRDY pin of only one device is used to detect new data.

0: Change to SPI_SYNC low.

1: Change to SPI_SYNC high.

#### SINGLE_SHOT_EN

One-shot mode. Enables one-shot mode. In one-shot mode, the AD7768-4
output a conversion result in response to a SYNC_IN rising edge.
Disabled.
Enabled.

#### SPI_RESET

Soft reset. These bits allow a full device reset over the SPI port. **Two** successive commands must be received in the correct order to generate a reset: first, write **0x03** to the soft reset register, and then write **0x02** to the soft reset register. This sequence causes the digital core to reset and all registers return to their default values. Following a soft reset, if the SPI master sends a command to the AD7768-4, the devices respond on the
next frame to that command with an output of **0x0E00**.

00: No effect.

01: No effect.

10: Second reset command.

11: First reset command.

---

### INTERFACE CONFIGURATION

#### CRC_SELECT

CRC select. These bits allow the user to implement a CRC on the data interface. When selected, the CRC replaces the header every fourth or 16th output sample depending on the CRC option chosen. There are two options for the CRC; both use the same polynomial: x 8 + x 2 + x + 1. The options offer the user the ability to reduce the duty cycle of the CRC calculation by performing it less often: in the case of having it every 16th sample or more often in the case of every fourth conversion. The CRC is calculated on a per channel basis and it includes conversion data only.

00: No CRC. Status bits with every conversion.

01: Replace the header with CRC message every 4 samples.

10: Replace the header with CRC message every 16 samples.

11: Replace the header with CRC message every 16 samples.

#### DCLK_DIV

DCLK divider. These bits control division of the DCLK clock used to clock out conversion data on the DOUTx pins. The DCLK signal is derived from the MCLK applied to the AD7768-4. The DCLK divide mode allows the user to optimize the DCLK output to fit the application. Optimizing the DCLK per application depends on the requirements of the user. When the AD7768-4 are using the highest capacity output on the fewest DOUTx pins, for example, running in decimate by 32 using the DOUT0 and DOUT1 pins, the DCLK must equal the MCLK; thus, in this case, choosing the no division setting is the only way the user can output all the data within the conversion period. There are other cases, however, when the ADC may be running in fast mode with high decimation rates, or in median or low power mode where the DCLK does not need to run at the same speed as MCLK. In these cases, the DCLK divide allows the user to reduce the clock speed and makes routing and isolating such signals easier.

00: Divide by 8.

01: Divide by 4.

10: Divide by 2.

11: No division.
