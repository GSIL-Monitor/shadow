<?php
/**
 * Created by PhpStorm.
 * User: zhaoxi (zhaoxi@mogujie.com)
 * Date: 15-03-05
 * Time: 15:02
 */
require __DIR__ . '/format.php';
require __DIR__ . '/config.php';

$host = $config['host'];
$port = $config['port'];


$keySize = intval($argv[1]);
$valueSize = intval($argv[2]);
$maxDuration = intval($argv[3]); // second. system running time


echo "keySize: ". $keySize . "\n";
echo "valueSize: ". $valueSize . "\n";
echo "maxDuration: ". $maxDuration . "\n";

run($size, $host, $port, $keySize, $valueSize, $maxDuration);

function run($size, $host, $port, $keySize, $valueSize, $maxDuration) {
	$duration = 0;
	do {
		$start = time();

		$key = buildValue($keySize);
		$value = buildValue($valueSize);
		write( $host, $port, $key, $value);
		read( $host, $port, $key, $value);

		$duration += time() - $start;
	} while ($duration <= $maxDuration);
}

function write( $host, $port, $key , $value ) {
	$errFile = "/tmp/error.log";
	try {

		$redis = new Redis();
		$redis->connect($host, $port, 5);
		$resp = $redis->setex($key, 1000 * 60 * 5, $value);	
		if( !$resp ) {
			echo "===";
		}
		$redis->close();
		unset( $redis );
	} catch( Exception $e) {
		$content = var_export($e, true) . "\n";
		file_put_contents($errFile, $content, FILE_APPEND);
	}
}

function read( $host, $port, $key, $value ) {
	$errFile = "/tmp/error.log";
	try {
		$redis = new Redis();
		$redis->connect($host, $port, 5);
		$resp = $redis->get($key);	
		$redis->close();
		unset( $redis );
		if( $resp !== null && strcmp( $value, $resp) === 0 ) {
			// It is ok.
		} else {
			echo "***" . $key;
			$content = "I can not get expected value. The key is $key. \n";
			echo "***" . $content;
			file_put_contents($errFile, $content, FILE_APPEND);
		}	
	} catch( Exception $e) {
		$content = var_export($e, true) . "\n";
		file_put_contents($errFile, $content, FILE_APPEND);
	}
}
