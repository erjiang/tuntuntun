#include <sys/socket.h>
#include <fcntl.h>
#include <linux/if.h>
#include <linux/if_tun.h>
#include <stdio.h>
#include <string.h>
#include <unistd.h>
#include <sys/ioctl.h>

#include "readpacket.h"

#define BUF_SIZE (2048)

// original main function for testing
int old_main(int argc, char **argv) {
    char readbuf[BUF_SIZE];
    char tun_name[IFNAMSIZ];
    strcpy(tun_name, "tun0");
    printf("Creating tun...\n");
    int my_tun = tun_alloc(tun_name, IFF_TUN | IFF_NO_PI);
    if (my_tun < 0) {
        printf("Could not create tun0. Permissions not correct?\n");
        _exit(-1);
    }
    printf("Listening on fd %d...\n", my_tun);
    
    for(;;) {
        ssize_t b = read(my_tun, readbuf, sizeof readbuf);
        print_bytes(readbuf, b);
    }
}

char* alloc_tun_name() {
    char *nam = malloc(IFNAMSIZ);
    bzero(nam, IFNAMSIZ);
    return nam;
}

void print_bytes(char *bytes, ssize_t n) {
    int i = 0;
    while (i < n) {
        printf("%02hhx ", bytes[i]);
        i++;
        if (i % 20 == 0) {
            printf("\n");
        }
    }
}

int tun_alloc(char *dev, int flags) {

  struct ifreq ifr;
  int fd, err;
  char *clonedev = "/dev/net/tun";

  /* Arguments taken by the function:
   *
   * char *dev: the name of an interface (or '\0'). MUST have enough
   *   space to hold the interface name if '\0' is passed
   * int flags: interface flags (eg, IFF_TUN etc.)
   */

   /* open the clone device */
   if( (fd = open(clonedev, O_RDWR)) < 0 ) {
     return fd;
   }

   /* preparation of the struct ifr, of type "struct ifreq" */
   memset(&ifr, 0, sizeof(ifr));

   ifr.ifr_flags = flags;   /* IFF_TUN or IFF_TAP, plus maybe IFF_NO_PI */

   if (*dev) {
     /* if a device name was specified, put it in the structure; otherwise,
      * the kernel will try to allocate the "next" device of the
      * specified type */
     strncpy(ifr.ifr_name, dev, IFNAMSIZ);
   }

   /* try to create the device */
   if( (err = ioctl(fd, TUNSETIFF, (void *) &ifr)) < 0 ) {
     close(fd);
     return err;
   }

  /* if the operation was successful, write back the name of the
   * interface to the variable "dev", so the caller can know
   * it. Note that the caller MUST reserve space in *dev (see calling
   * code below) */
  strcpy(dev, ifr.ifr_name);

  /* this is the special file descriptor that the caller will use to talk
   * with the virtual interface */
  return fd;
}
