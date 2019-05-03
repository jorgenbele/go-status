#!/bin/awk -f
# Awk script to convert nmcli output
# to i3bar block.
BEGIN { 
   FS = ":";
   color = "#FFFFFF" 

   if (ARGC == 2) { n = ARGV[1] } 

   command = "nmcli -g CONNECTION,STATE dev status"

   printf "["

   while ((n == 0 || count < n) && (command |& getline) > 0) {
       if (length($1) == 0) {
        continue;
       }

       if (count > 0) 
        printf ","
       count += 1

       if ($2 == "connected")         color = "#888888";
       else if ($2 == "disconnected") color = "#FF0000";

       printf "{\"full_text\": \"%s\", \"align\": \"%s\", \"name\": \"%s\", \"color\": \"%s\"}\n", $1, "right", "nmcli", color;
   }
   printf "]"

   close(command)
}
