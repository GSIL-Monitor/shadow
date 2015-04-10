<?php
/**
 * Created by PhpStorm.
 * User: zhaoxi (zhaoxi@mogujie.com)
 * Date: 15-03-05
 * Time: 15:02
 */
require __DIR__ . '/../shadow/cache.php';
require __DIR__ . '/format.php';
$app = 'StressTesting';


$size = 0;
$index = 0;
$maxDuration = 0;
$r = 1;
$w = 1;

$size = intval($argv[1]);
$index = intval($argv[2]);
$maxDuration = intval($argv[3]);
$r = intval($argv[4]);
$w = intval($argv[5]);


testShadow($app, $size, $r, $w, $index, $maxDuration);

function testShadow($app, $size, $readTime, $writeTime, $index, $maxDuration) {
    $resp = array();
    $records = array(); // $cost => $count
    $resp["query"] = 0;
    $resp['cost_ms/10'] = 0;
    $resp['avg_cost_ms'] = 0;
    $resp["set_ok"] = 0;
    $resp["get_ok"] = 0;


    generate($app, $size, $readTime, $writeTime, $resp, $records, $maxDuration);
    // file
    writeSummary($index, $resp, $size);
    writeStatus($index, $records);
}


function generate(&$servers, $size, $readTime, $writeTime, &$resp, &$records, $maxDuration) {
    $c = $readTime + $writeTime;
    $duration = 0;
    do {
        $cost = 0;
        $start = time();
        run($servers, $size, $readTime, $writeTime, $resp, $cost);
        $duration += time() - $start;
        $resp["query"] = $resp["query"] + 2;
        $resp['cost_ms/10'] += $cost;
        $resp['avg_cost_ms'] = ($resp['cost_ms/10']) / (10.0 * $resp["query"]);
        // status
        $k = $cost . "";  // =>  $count
        if (isset($records[$k])) {
            $records[$k] = $records[$k] + $c;
        } else {
            $records[$k] = $c;
        }
    } while ($duration <= $maxDuration);
}


function run($app, $size, $readTime, $writeTime, &$resp, &$cost) {
    $errFile = "/tmp/error.log";
    $key = buildKey(10);
    $value = build($size);
    $cache = Shadow_Cache::build($app);
    for ($i = 0; $i < $writeTime; $i++) {
        try {
            // set
            $start = microtime(true);
            $r = $cache->setex($key, 600, $value);
            $cost += cost(microtime(true), $start);
            if ($r) {
                $resp["set_ok"] += 1;
            }
        } catch (Exception $e) {
            $date = new DateTime('now');
            $s = $date->format('Y-m-d H:i:s');
            $content = $s . "\n" . var_export($e, true) . ";\n";
            file_put_contents($errFile, $content, FILE_APPEND);
        }
    }


    for ($i = 0; $i < $readTime; $i++) {
        try {
            // get
            $start = microtime(true);
            $r = $cache->get($key);
            $cost += cost(microtime(true), $start);
            if (strcmp($r, $value) == 0) {
                $resp["get_ok"] += 1;
            }
        } catch (Exception $e) {
            $date = new DateTime('now');
            $s = $date->format('Y-m-d H:i:s');
            $content = $s . "\n" . var_export($e, true) . ";\n";
            file_put_contents($errFile, $content, FILE_APPEND);
        }
    }

    $cache = null;
}


function writeSummary($index, &$resp, $size) {
    $f = "/tmp/shadow/" . $size . "_bytes_shadow_status_" . $index . ".txt";
    $content = "";
    $cnt = count($resp) - 1;
    // header
    $i = 0;
    foreach ($resp as $k => $v) {
        $content .= $k;
        if ($i < $cnt) {
            $content .= "\t";
        }
        $i++;
    }
    $content .= "\n";

    // body
    $i = 0;
    foreach ($resp as $k => $v) {
        $content .= $v;
        if ($i < $cnt) {
            $content .= "\t";
        }
        $i++;
    }
    $content .= "\n";
    file_put_contents($f, $content);
}

function writeStatus($index, &$records) {
    uksort($records, "byCost");
    foreach ($records as $cost => $count) {
        echo $cost . "," . $count . "\n";
    }
}

function byCost($k1, $k2) {
    return doubleval($k1) - doubleval($k2) > 0;
}