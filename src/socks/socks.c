#include <netinet/in.h>
#include <errno.h>
#include <stdlib.h>
#include <unistd.h>
#include <netdb.h>
#include <string.h>
#include <stdio.h>
#include "socks.h"

int createDeviceBoundUDPSocket(uint32_t sip, uint16_t sport, const char* bind_dev) {
    printf("bind_dev = %s", bind_dev);
   int s = socket(AF_INET, SOCK_DGRAM, 0);   
   int result;
   struct sockaddr_in my_ip_addr;

   if (s < 0) {
      perror("socket");
      return s;
   }

   memset(&my_ip_addr, 0, sizeof(my_ip_addr));
   
   my_ip_addr.sin_family = AF_INET;
   my_ip_addr.sin_addr.s_addr = htonl(sip);
   my_ip_addr.sin_port = htons(sport);
   
   /*
   result = bind(s, (struct sockaddr*)(&my_ip_addr), sizeof(my_ip_addr));
   if (result < 0) {
      perror("Error in bind");
      return result;
   }
   */
   
   if (bind_dev) {
      // Bind to specific device.
      if (setsockopt(s, SOL_SOCKET, SO_BINDTODEVICE,
                     bind_dev, strlen(bind_dev) + 1)) {
         perror("Error binding to device");
         return -1;
      }
   }

   return s;
}

ssize_t writeToUDP(int fd, void *buf, size_t buflen, uint32_t dip, uint16_t dport) {
    putchar('X');
    struct sockaddr_in sa;
    struct in_addr ipa;
    sa.sin_addr.s_addr = htonl(dip);
    sa.sin_family = AF_INET;
    sa.sin_port = htons(dport);

    printf("Sending out %d bytes on fd %d to %x:%d", buflen, fd, dip, dport);
    return sendto(fd, buf, buflen, 0, &sa, sizeof(sa));
}
