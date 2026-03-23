Poltergeist is a evil spirit from German Folk lore in which it had invisible hands and it used it to control/destroy objects

this project is named so as i am trying to make it so that i can use my OnePlus Ear buds to control my laptop

yup i will be using the touch pad that is on the side of the bud and try to manipulate it's network packet in order to link them to my linux machines either gyroscope/screen

what i have tried till now : 
steps
1. get adb to read bugreport from my BLE/Bluetooth
2. used btsnooz.py but it expected other formats
4. changed format
5. got corrupted data
6. trying btmon - allows to read bluetooth HCI
7. got btmon1.log and btmon2.log in /FS/data/misc/bluetooth/logs
8. got to learn about the wireshark software
9. all early efforts were for naught
10. converted .cfa files to .pcap...named output1.pcap
11. listened on the bluetooth monitor interface
12. found mac address of my ear buds as
13. bluetooth.addr == 9c:de:f0:33:f5:dd
14. got to know that wireshark was only capturing the classic bluetooth (BR/EDR) not BLE for GATT
15. streaming audio via A2DP (AVDTP + RTP)
16. controlling via AVRCP
17. got BLE gatt from bugreport




stay tuned for more :D
