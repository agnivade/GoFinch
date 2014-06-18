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
func (finch *Finch) incrementSequenceNumber() {
  if finch.sequence_number+1 > 255 {
    finch.sequence_number = 0
  } else {
    finch.sequence_number++
  }
}

//------------------------------------------------------------------------------
func (finch *Finch) writeToFinch(data []byte) (n int, err error) {
  n = 0
  // Writing until we see that we have atleast written something
  for n==0 {
    n, err = finch.finch_handle.Write(data)
  }
  return
}

//------------------------------------------------------------------------------
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
func (finch *Finch) TurnOffMotorAndLEDs() (n int, err error) {
  data := prepareFinchRequest()
  data[1] = 'X'

  n, err = finch.writeToFinch(data)
  return
}

//------------------------------------------------------------------------------
func (finch *Finch) SetIdleMode() (n int, err error) {
  data := prepareFinchRequest()
  data[1] = 'R'

  n, err = finch.writeToFinch(data)
  return
}

//------------------------------------------------------------------------------
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

  if bit_5 == 0 {
    tap = true
  }
  if bit_7 == 1 {
    shake = true
  }
  return
}

//------------------------------------------------------------------------------
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
func (finch *Finch) Close() () {
  // First going to idle mode
  finch.SetIdleMode()
  // Then closing the handle to finch
  finch.finch_handle.Close()
}

