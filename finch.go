// This is a Go implementation of the Finch API
// For more information - visit http://www.finchrobot.com
package finch

import (
  "fmt"
  "time"
  "github.com/GeertJohan/go.hid"
  )
// TODO: change the return types to use structs instead of
// returning individual values

type Finch struct {
  finch_handle *hid.Device
  sequence_number byte
  read_timeout_msec int
}

//------------------------------------------------------------------------------
// Internal helper function to prepare the 9 bytes array
// Set all the bits to 0 and then return
func prepareFinchRequest() (data []byte) {
  data = make([]byte, FINCH_MESSAGE_LENGTH_BYTES)
  for i:=0; i<FINCH_MESSAGE_LENGTH_BYTES; i++ {
    data[i] = 0x00
  }
  return
}

//------------------------------------------------------------------------------
// Increments the sequence no. of the request to finch
// Rolls over to 0 if it exceeds 255
func (finch *Finch) incrementSequenceNumber() {
  if finch.sequence_number+1 > 255 {
    finch.sequence_number = 0
  } else {
    finch.sequence_number++
  }
}

//------------------------------------------------------------------------------
// Helper function which writes to finch given a byte slice
func (finch *Finch) writeToFinch(data []byte) (n int, err error) {
  n = 0
  // Writing until we see that we have atleast written something
  for n==0 {
    n, err = finch.finch_handle.Write(data)
  }
  return
}

//------------------------------------------------------------------------------
// Helper function which reads from finch and fills a byte slice
func (finch *Finch) readFromFinch(data []byte) (n int, err error) {
  n = len(data)
  if n <= 0 {
    return 0, nil
  }

  // Reading until there is nothing else to read
  for n > 0 {
    n, err = finch.finch_handle.ReadTimeout(data, finch.read_timeout_msec)
    // Breaking if we see that the sequence no. matches
    if data[7] == data[8] {
      break
    }
    if err != nil {
      return
    }
  }
  return
}

//------------------------------------------------------------------------------
// Open connects to the Finch device and returns a Finch struct
func Open() (finch *Finch, err error) {
  var device_handle *hid.Device
  device_handle, err = hid.Open(VENDOR_ID, DEVICE_ID, "")
  if err != nil {
    return nil, err
  }

  finch = &Finch{
    finch_handle: device_handle,
    sequence_number: 0,
    read_timeout_msec: 1*1000, // 1 second
  }
  return
}

//------------------------------------------------------------------------------
// SetLed sets the LED of the Finch. You can specify the red, green and blue
// intensities from the params in range of 0-255.
func (finch *Finch) SetLed(red, green, blue byte) (n int, err error) {
  data := prepareFinchRequest()
  data[1] = 'O'   // Ascii character 'O' for LED color set
  data[2] = red   // value of red
  data[3] = green // value of green
  data[4] = blue  // value of blue

  n, err = finch.writeToFinch(data)
  return
}

//------------------------------------------------------------------------------
// SetMotor sets the motor speed and wheel direction of the Finch. The direction
// must be either 0 or 1. 0 meaning forward and 1 meaning reverse. The speed
// must be within 0 to 255.
func (finch *Finch) SetMotor(left_wheel_direction,
                                  left_wheel_speed,
                                  right_wheel_direction,
                                  right_wheel_speed byte) (n int, err error) {
  if left_wheel_direction != 0 && left_wheel_direction != 1 {
    err := fmt.Errorf("The value of left wheel direction must lie be either 0 or 1. You have passed- %d", left_wheel_direction)
    return 0, err
  }

  if right_wheel_direction != 0 && right_wheel_direction != 1 {
    err := fmt.Errorf("The value of right wheel direction must lie be either 0 or 1. You have passed- %d", right_wheel_direction)
    return 0, err
  }

  data := prepareFinchRequest()
  data[1] = 'M'   // Ascii character 'O' for LED color set
  data[2] = left_wheel_direction   // value of red
  data[3] = left_wheel_speed // value of green
  data[4] = right_wheel_direction  // value of blue
  data[5] = right_wheel_speed  // dummy data

  n, err = finch.writeToFinch(data)
  return
}

//------------------------------------------------------------------------------
// TurnOffMotorAndLEDs sets all motors and LEDs to off.
func (finch *Finch) TurnOffMotorAndLEDs() (n int, err error) {
  data := prepareFinchRequest()
  data[1] = 'X'

  n, err = finch.writeToFinch(data)
  return
}

//------------------------------------------------------------------------------
// SetIdleMode turns off the motor and has the Finch go back to
// color-cycling mode
func (finch *Finch) SetIdleMode() (n int, err error) {
  data := prepareFinchRequest()
  data[1] = 'R'

  n, err = finch.writeToFinch(data)
  return
}

//------------------------------------------------------------------------------
// SetBuzzer sets the buzzer to chirp for a specified period of time.
// There's an additional wait parameter which will block the function call
// for the time period it is set to chirp. Otherwise it will return
// immediately.
func (finch *Finch) SetBuzzer(msec, freq int, wait bool) (n int, err error) {
  data := prepareFinchRequest()
  data[1] = 'B'   // Ascii character 'O' for LED color set
  data[2] = byte(msec >> 8)   // value of red
  data[3] = byte(msec) // value of green
  data[4] = byte(freq >> 8)  // value of blue
  data[5] = byte(freq)  // dummy data

  n, err = finch.writeToFinch(data)

  if wait {
    time.Sleep(time.Duration(msec)*time.Millisecond)
  }
  return
}

//------------------------------------------------------------------------------
// GetTemperature returns the temp in celsius.
func (finch *Finch) GetTemperature() (temp float64, err error) {
  finch.incrementSequenceNumber()

  data := prepareFinchRequest()
  data[1] = 'T'   // Ascii character 'O' for LED color set
  data[8] = finch.sequence_number  // dummy data

  _, err = finch.writeToFinch(data)
  if err != nil {
      return 0, err
  }

  _, err = finch.readFromFinch(data)
  if err != nil {
      return 0, err
  }

  temp = float64(int(data[0]) - 127)/2.4 + 25
  return
}

//------------------------------------------------------------------------------
// GetLight returns the values from the two light sensors. The values
// are in the range of 0-255
func (finch *Finch) GetLight() (left_sensor, right_sensor byte, err error) {
  finch.incrementSequenceNumber()

  data := prepareFinchRequest()
  data[1] = 'L'
  data[8] = finch.sequence_number

  _, err = finch.writeToFinch(data)
  if err != nil {
      return 0, 0, err
  }

  _, err = finch.readFromFinch(data)
  if err != nil {
    return 0, 0, err
  }

  left_sensor = data[0]
  right_sensor = data[1]
  return
}

//------------------------------------------------------------------------------
// GetAcceleration returns the accelerometer values of the x, y and z axis.
// And also the tap/shake byte
func (finch *Finch) GetAcceleration() (x_axis, y_axis, z_axis float64,
                                       tap, shake bool,
                                       err error) {
  finch.incrementSequenceNumber()
  tap = false
  shake = false

  data := prepareFinchRequest()
  data[1] = 'A'
  data[8] = finch.sequence_number

  _, err = finch.writeToFinch(data)
  if err != nil {
      return 0, 0, 0, false, false, err
  }

  _, err = finch.readFromFinch(data)
  if err != nil {
      return 0, 0, 0, false, false, err
  }

  // A closure to convert the raw machine data into g's
  convertToG := func(b byte) float64 {
    // This the calculation to convert the raw machine data to g
    if b > 31 {
      b -= 64
    }
    return float64(b)*1.5 / 32
  }
  x_axis = convertToG(data[1])
  y_axis = convertToG(data[2])
  z_axis = convertToG(data[3])

  bit_5 := data[4] & 0x20
  bit_7 := data[4] & 0x80

  // If bit 5 is a 0, then finch has been tapped since the last read
  if bit_5 == 0 {
    tap = true
  }
  // If bit 7 is a 1, then finch has been shaken since the last read
  if bit_7 == 1 {
    shake = true
  }
  return
}

//------------------------------------------------------------------------------
// GetObstacles returns the values of the two obstacle sensors. true means
// obstacle is present, and false means absent
func (finch *Finch) GetObstacles() (left_sensor, right_sensor bool, err error) {
  finch.incrementSequenceNumber()

  data := prepareFinchRequest()
  data[1] = 'I'
  data[8] = finch.sequence_number

  _, err = finch.writeToFinch(data)
  if err != nil {
      return false, false, err
  }

  _, err = finch.readFromFinch(data)
  if err != nil {
      return false, false, err
  }

  left_sensor = false
  right_sensor = false

  if data[0] == 0 {
    left_sensor = false
  } else if data[0] == 1 {
    left_sensor = true
  }

  if data[1] == 0 {
    right_sensor = false
  } else if data[1] == 1 {
    right_sensor = true
  }

  return
}

//------------------------------------------------------------------------------
// Close sets the Finch to idle mode and then closes the connection to it.
func (finch *Finch) Close() () {
  // First going to idle mode
  finch.SetIdleMode()
  // Then closing the handle to finch
  finch.finch_handle.Close()
}

