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

## delete
