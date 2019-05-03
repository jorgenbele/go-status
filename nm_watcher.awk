#!/bin/awk -f
# Author: Jørgen Bele Reinfjell
# Date: 03.05.2019 [dd.mm.yyyy]
# File: nm_watcher.awk
# Description:
#   Awk script to convert 'nmcli dev monitor' output
#   to i3bar block.

function reset_state()
{
    color     = "#FFFFFF"
    istatus   = "unknown"
    iconn     = "unknown"
    iconns    = ""
}

function initialstate(ifname)
{
    reset_state()
    command = "nmcli -g DEVICE,STATE,CONNECTION dev status"
    orig = FS
    FS       = ":"
    while ((command |& getline) > 0) {
        if (length($1) == 0) {
            continue;
        }

        if ($1 == ifname) {
            istatus = $2
            iconn = $3
            break
        }
    }
    close(command)
    FS = orig
}

function updatestate(ifname)
{
    if ($1 == ifname) {
        if ($2 ~ /using connection.*/)   { reset_state(); split($2, sp, "'"); iconn = sp[2]; }
        else if ($2 == "connected")      {                istatus="connected";    color = "#FFFFFF" }
        else if ($2 == "deactivating")   { reset_state(); istatus="deactivating"; color = "#FF5500" }
        else if ($2 ~ /connecting.*/)    {
            istatus = $2
            color = "#009900"
        }
        else if ($2 == "disconnected")   { reset_state(); istatus="disconnected"; color = "#FF0000" }
        else if ($2 == "device removed") {                istatus="removed";      color = "#FF0000" }
    }
}

function printstate()
{
    if (istatus != "connected")
        printf "{\"full_text\": \"%s: %s%s\", \"align\": \"%s\", \"name\": \"%s\", \"color\": \"%s\"}\n", iconn, istatus, iconns, "right", "nmcli_con", color;
    else
        printf "{\"full_text\": \"%s\", \"align\": \"%s\", \"name\": \"%s\", \"color\": \"%s\"}\n", iconn, "right", "nmcli_con", color;
    fflush()
}

BEGIN {
   FS = ": "

   if (ARGC == 2) ifname = ARGV[1]
   else { print "Usage: awk -f nm_watcher.awk IFNAME"; exit(1) }

   initialstate(ifname)
   printstate()

   for (;;) {
       command = "nmcli dev monitor"

       while ((command |& getline) > 0) {
           if (length($1) == 0) {
            continue;
           }
           updatestate(ifname)
           printstate()
       }
       close(command)
   }
}
