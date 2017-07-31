#ifndef GO_DHT_H
#define GO_DHT_H

#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <fcntl.h>
#include <sched.h>
#include <time.h>
#include <unistd.h>

// GPIO direction: receive either output data to specific GPIO pin.
#define IN  0
#define OUT 1
 
// LOW correspong to low level of output signal, HIGH correspond to high level.
#define LOW  0
#define HIGH 1

// TRUE, FALSE values
#define FALSE 0
#define TRUE 1

// Keep pin no, file descriptors for data reading/writing
// and for specifying input/output mode.
typedef struct {
    int pin;
    // Keep file descriptors for "direction" and "value"
    // open during whole sensor session interraction,
    // because it save milliseconds critical for one-wire
    // DHTxx protocol
    int fd_direction;
    int fd_value;
} Pin;
 
// Freeze thread for usec microseconds
static int sleep_usec(int32_t usec) {
    struct timespec tim, tim2;
    // convert microseconds to seconds
    tim.tv_sec = usec / 1000000;
    // rest part of microseconds convert to nanoseconds
    tim.tv_nsec = (usec % 1000000) * 1000;
    return nanosleep(&tim , &tim2);
}

// Start working with specific pin.
static int gpio_export(int port, Pin *pin) {
    #define BUFFER_MAX 3
    char buffer[BUFFER_MAX];
    ssize_t bytes_written;
    int fd;

    // initialize pin to work with, "direction" and "value"
    // file descriptors with empty value
    (*pin).pin = -1;
    (*pin).fd_direction = -1;
    (*pin).fd_value = -1;
                 
    fd = open("/sys/class/gpio/export", O_WRONLY);
    if (-1 == fd) {
        fprintf(stderr, "Failed to open export for writing!\n");
        return -1;
    }
    (*pin).pin = port;
    bytes_written = snprintf(buffer, BUFFER_MAX, "%d", (*pin).pin);
    if (-1 == write(fd, buffer, bytes_written)) {
        fprintf(stderr, "Failed to export pin!\n");
        close(fd);
        return -1;
    }
    close(fd);

    // !!! Found in experimental way, that additional pause should exist
    // between export pin to work with and direction set up. Otherwise,
    // under the regular user mistake occures frequently !!!
    //
    // Sleep 150 milliseconds
    sleep_usec(150*1000);

    #define DIRECTION_MAX 35
    char path1[DIRECTION_MAX];
    snprintf(path1, DIRECTION_MAX, "/sys/class/gpio/gpio%d/direction", (*pin).pin);
    (*pin).fd_direction = open(path1, O_WRONLY);
    if (-1 == (*pin).fd_direction) {
        fprintf(stderr, "Failed to open gpio direction for writing!\n");
        return -1;
    }
                             
    #define VALUE_MAX 30
    char path2[VALUE_MAX];
    snprintf(path2, VALUE_MAX, "/sys/class/gpio/gpio%d/value", (*pin).pin);
    (*pin).fd_value = open(path2, O_RDWR);
    if (-1 == (*pin).fd_value) {
        fprintf(stderr, "Failed to open gpio value for reading!\n");
        return -1;
    }
                             
    return 0;
}

// Stop working with specific pin.
static int gpio_unexport(Pin *pin) {
    // close "direction" file descriptor
    if (-1 != (*pin).fd_direction) {
        close((*pin).fd_direction);
        (*pin).fd_direction = -1;
    }
    // close "value" file descriptor
    if (-1 != (*pin).fd_value) {
        close((*pin).fd_value);
        (*pin).fd_value = -1;
    }

    if (-1 != (*pin).pin) {
        char buffer[BUFFER_MAX];
        ssize_t bytes_written;
        int fd;
                 
        fd = open("/sys/class/gpio/unexport", O_WRONLY);
        if (-1 == fd) {
            fprintf(stderr, "Failed to open unexport for writing!\n");
            return -1;
        }
                         
        bytes_written = snprintf(buffer, BUFFER_MAX, "%d", (*pin).pin);
        if (-1 == write(fd, buffer, bytes_written)) {
            fprintf(stderr, "Failed to unexport pin!\n");
            close(fd);
            return -1;
        }

        close(fd);
    }
    return 0;
}
 
// Setup pin mode: input or output.
static int gpio_direction(Pin *pin, int dir) {
    static const char s_directions_str[]  = "in\0out";
         
    if (-1 == write((*pin).fd_direction, &s_directions_str[IN == dir ? 0 : 3], IN == dir ? 2 : 3)) {
        fprintf(stderr, "Failed to set direction!\n");
        return -1;
    }
    return 0;
}

// Read data from the pin: in normal conditions return 0 or 1,
// which correspond to low or high signal levels.
static int gpio_read(Pin *pin) {
    char value_str[3];
                 
    if (-1 == lseek((*pin).fd_value, 0, SEEK_SET)) {
        fprintf(stderr, "Failed to seek file!\n");
        return -1;
    }
    if (-1 == read((*pin).fd_value, value_str, 3)) {
        fprintf(stderr, "Failed to read value!\n");
        return -1;
    }

    // Small optimization to speed up GPIO processing
    // due to ARM devices CPU slowness.
    if (value_str[1] == '\0') {
        return value_str[0] == '0' ? 0 : 1;
    } else {
        return atoi(value_str);
    }
}

// Macro to convert timespec structure value to microseconds.
#define convert_timespec_to_usec(t) ((t).tv_sec*1000*1000 + (t).tv_nsec/1000)
 
// Read sequence of data from the pin trigering
// on edge change until timeout occures.
// Collect as well durations of pulses in microseconds.
// Fill [arr] array with a sequence: level1, duration1, level2, duration2...
// Put array length to variable [len].
static int gpio_read_seq_until_timeout(Pin *pin,
        int32_t timeout_msec, int32_t **arr, int32_t *len) {
    int32_t last_v, next_v;
#define MAX_PULSE_COUNT 16000
    int values[MAX_PULSE_COUNT*2];
    
    last_v = gpio_read(pin);
    if (-1 == last_v) {
        fprintf(stderr, "Failed to read value!\n");
        return -1;
    }
    int k = 0, i = 0;
    values[k*2] = last_v;
    struct timespec last_t, next_t;
#define CLOCK_KIND CLOCK_MONOTONIC
// #define CLOCK_KIND CLOCK_REALTIME
    clock_gettime(CLOCK_KIND, &last_t);

    for (;;)
    {
        // sleep_microsec(1);
        next_v = gpio_read(pin);
        if (-1 == next_v) {
            fprintf(stderr, "Failed to read value!\n");
            return -1;
        }

        if (last_v != next_v) {
            clock_gettime(CLOCK_KIND, &next_t); 
            i = 0;
            k++;
            if (k > MAX_PULSE_COUNT-1) {
                fprintf(stderr, "Pulse count exceed limit in %d\n", MAX_PULSE_COUNT);
                return -1;
            }
            values[k*2] = next_v;
            // Save time duration in microseconds.
            values[k*2-1] = convert_timespec_to_usec(next_t) -
                convert_timespec_to_usec(last_t); 
            last_v = next_v;
            last_t = next_t;
        }

        if (i++ > 20) {
            clock_gettime(CLOCK_KIND, &next_t); 
            if ((convert_timespec_to_usec(next_t) -
                convert_timespec_to_usec(last_t)) / 1000 > timeout_msec) {
                values[k*2+1] = timeout_msec * 1000;
                break;
            }
        }
    }
    *arr = malloc((k+1)*2 * sizeof(int32_t));
    for (i=0; i<=k; i++)
    {
        (*arr)[i*2] = values[i*2];
        (*arr)[i*2+1] = values[i*2+1];
    }
    *len = (k+1)*2;
                                 
/*    fprintf(stdout, "scan %d values\n", k+1);
    for (i=0; i<=k; i++)
    {
        fprintf(stdout, "value %d (%d): %d\n", i, (*arr)[i*2+1], (*arr)[i*2]);
    }*/
    return 0;
}
 
// Set up specific pin level to 0 (low) or 1 (high).
static int gpio_write(Pin *pin, int value) {
    static const char s_values_str[] = "01";
         
    if (1 != write((*pin).fd_value, &s_values_str[LOW == value ? 0 : 1], 1)) {
        fprintf(stderr, "Failed to write value!\n");
        return -1;
    }
                                 
    return 0;
}

// Used to gain maximum performance from device during
// receiving bunch of data from sensors like DHTxx.
static int set_max_priority(void) {
    struct sched_param sched;
    memset(&sched, 0, sizeof(sched));
    // Use FIFO scheduler with highest priority
    // for the lowest chance of the kernel context switching.
    sched.sched_priority = sched_get_priority_max(SCHED_FIFO);
    if (-1 == sched_setscheduler(0, SCHED_FIFO, &sched)) {
        fprintf(stderr, "Unable to set SCHED_FIFO priority to the thread\n");
        return -1;
    }
    return 0;
}

// Get back normal thread priority.
static int set_default_priority(void) {
    struct sched_param sched;
    memset(&sched, 0, sizeof(sched));
    // Go back to regular schedule priority.
    sched.sched_priority = 0;
    if (-1 == sched_setscheduler(0, SCHED_OTHER, &sched)) {
        fprintf(stderr, "Unable to set SCHED_OTHER priority to the thread\n");
        return -1;
    }
    return 0;
}

typedef struct {
    int time;
    int edge;
} edge_info;


// Blink specific pin n times. Led could be
// attached to this pin for debug purpose.
static int blink_n_times(int pin, int n) {
    Pin p;
    if (-1 == gpio_export(pin, &p)) {
        gpio_unexport(&p);
        return -1;
    }
    if (-1 == gpio_direction(&p, OUT)) {
        gpio_unexport(&p);
        return -1;
    }
    int i;
    // Blink led n times in a loop.
    for (i = 0; i < n; i++)
    {
        // Turn led on
        if (-1 == gpio_write(&p, HIGH)) {
            gpio_unexport(&p);
            return -1;
        }
        // Sleep 0.1 of second.
        sleep_usec(100*1000);
        // Turn led off
        if (-1 == gpio_write(&p, LOW)) {
            gpio_unexport(&p);
            return -1;
        }
        // Sleep 0.1 of second.
        sleep_usec(100*1000);
    }
    return gpio_unexport(&p);
}

// Activate DHTxx sensor and collect data sent by sensor for futher processing.
static int dial_DHTxx_and_read(int32_t pin, int32_t boostPerfFlag,
        int32_t **arr, int32_t *arr_len) {
    // Set maximum priority for GPIO processing.
    if (boostPerfFlag != FALSE && -1 == set_max_priority()) {
        return -1;
    }
    Pin p;
    if (-1 == gpio_export(pin, &p)) {
        gpio_unexport(&p);
        set_default_priority();
        return -1;
    }
    // Send dial pulse.
    if (-1 == gpio_direction(&p, OUT)) {
        gpio_unexport(&p);
        set_default_priority();
        return -1;
    }
    // Set pin to high.
    if (-1 == gpio_write(&p, HIGH)) {
        gpio_unexport(&p);
        set_default_priority();
        return -1;
    }
    // Sleep 500 millisecond.
    sleep_usec(500*1000); 
    // Set pin to low.
    if (-1 == gpio_write(&p, LOW)) {
        gpio_unexport(&p);
        set_default_priority();
        return -1;
    }
    // Sleep 18 milliseconds according to DHTxx specification.
    sleep_usec(18*1000); 
    // Switch pin to input mode
    if (-1 == gpio_direction(&p, IN)) {
        gpio_unexport(&p);
        set_default_priority();
        return -1;
    }
    // Read bunch of data from sensor
    // for futher processing in high level language.
    // Wait for next pulse 10ms maximum.
    if (-1 == gpio_read_seq_until_timeout(&p, 10, arr, arr_len)) {
        gpio_unexport(&p);
        set_default_priority();
        return -1;
    }
    // Release pin.
    if (-1 == gpio_unexport(&p)) {
        set_default_priority();
        return -1;
    }
    // Return normal thread priority.
    if (boostPerfFlag != FALSE && -1 == set_default_priority()) {
        return -1;
    }
    return 0;
}

#endif
