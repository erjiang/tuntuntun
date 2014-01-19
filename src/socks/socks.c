#include <netinet/in.h>
#include <errno.h>
#include <stdlib.h>
#include <unistd.h>
#include <netdb.h>
#include <string.h>
#include <stdio.h>

int createDeviceBoundUDPSocket(uint32_t sip, uint16_t sport, const char* bind_dev) {
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
   
   result = bind(s, (struct sockaddr*)(&my_ip_addr), sizeof(my_ip_addr));
   if (result < 0) {
      perror("Error in bind");
      return result;
   }
   
   if (bind_dev) {
      // Bind to specific device.
      if (setsockopt(s, SOL_SOCKET, SO_BINDTODEVICE,
                     bind_dev, strlen(bind_dev) + 1)) {
         perror("Error binding to device");
         return -1;
      }
   }//if 

   return s;
}
