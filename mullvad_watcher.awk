#!/bin/awk -f
# Author: Jørgen Bele Reinfjell
# Date: 03.05.2019 [dd.mm.yyyy]
# File: mullvad_watcher.awk
# Description:
#   Awk script to convert 'mullvad status listen' output
#   to i3bar block.

function reset_state()
{
    color    = "#FFFFFF"
    tstatus  = "unknown"
    relay    = "unknown"
    location = "unknown"
    position = "unknown"

    icon     = " "
}

function updatestate()
{
    if ($1 == "Tunnel status") {
        if (match($2, "^Connected.*"))           { relay = "connected"; tstatus="connected"; color = "#FFFFFF" }
        else if ($2 == "Disconnecting...")       { reset_state(); tstatus="disconnecting";   color = "#FF5500" }
        else if ($2 == "Disconnected")           { reset_state(); tstatus="disconnected";    color = "#FF5500" }
        else if (match($2, "^Connecting.*"))     {                tstatus="connecting";      color = "#009900" }
        else if ($2 == "Blocked")                { reset_state(); tstatus="blocked";         color = "#FF0000" }
    }
    else if ($1 == "Relay")    relay    = $2
    else if ($1 == "Location") location = $2
    else if ($1 == "Position") position = $2
}

function printstate()
{
    if (tstatus != "connected")
        printf "{\"full_text\": \"%s%s\", \"align\": \"%s\", \"name\": \"%s\", \"color\": \"%s\"}\n", icon, tstatus, "right", "nmcli_con", color;
    else
        printf "{\"full_text\": \"%s%s\", \"align\": \"%s\", \"name\": \"%s\", \"color\": \"%s\"}\n", icon, relay, "right", "nmcli_con", color;
    fflush()
}

BEGIN {
   FS = ": "

   for (;;) {
       command = "mullvad status listen"
       reset_state()

       while ((command |& getline) > 0) {
           if (length($1) == 0) {
            continue;
           }

           updatestate()
           #printf "Status: %s, Relay: %s, Location: %s, Position: %s\n", tstatus, relay, location, position
           printstate()
       }
       close(command)
   }
}
