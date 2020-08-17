#!/usr/bin/php
<?php

if ($argc != 2) {
  print "rfepoch <date>\n";
  print "Returns RainForest time since epoch\n";
  print "<date> - time in UTC as yyyy-mm-dd hh:mm:ss\n";
  exit(1);
}

date_default_timezone_set('UTC');

$rfbase = 946684800;
$tparam = strtotime($argv[1]);
$delta = $tparam - $rfbase;

printf("Decimal %d\n", $delta);
printf("Hex %x\n", $delta);
exit(0);

?>
