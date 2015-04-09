<?php
/**
 * Created by PhpStorm.
 * User: zhaoxi (zhaoxi@mogujie.com)
 * Date: 15-02-27
 * Time: 16:49
 */
include('format.php');

$size = 0; // KB
$docker = 1; // 1 is docker, 0 is non-docker.

if (!isset($argv[1])) {
    echo "Please input the value size. Unit: KB.\n";
    exit;
}
if (!isset($argv[2])) {
    echo "Please input isDocker : 0 or 1.\n";
    exit;
}

$size = intval($argv[1]);
$docker = intval($argv[2]);


// ${query}, ${cost}, ${avg_cost}, ${set_ok}, ${get_ok}
// ${cost}, ${count}           Y:count, X:cost

if ($docker == 1) {
    testDocker($size);
} else if ($docker == 0) {
    testNonDocker($size);
}


function testNonDocker($size) {
    $servers = array();
    $s = array(
        "host" => "192.168.5.155",
        "port" => 6379,
    );
    $servers[] = $s;
    $key = buildKey(10);
    $value = buildValue($size);
    $resp = array();
    $records = array(); // $cost => $count
    $resp["query"] = 0;
    $resp['cost_ms/10'] = 0;
    $resp['avg_cost_ms'] = 0;
    $resp["set_ok"] = 0;
    $resp["get_ok"] = 0;

    $d = 0;
    generate($servers, $key, $value, $resp, $records, $d);
    // file
    writeSummary($resp, $size, 0);
    writeStatus($records);
}

/**
 * @param $size
 */
function testDocker($size) {
    $servers = array();
    $s = array(
        "host" => "192.168.5.160",
        "port" => 6380,
    );
    $servers[] = $s;
    $key = buildKey(10);
    $value = buildValue($size);
    $resp = array();
    $records = array(); // $cost => $count
    $resp["query"] = 0;
    $resp['cost_ms/10'] = 0;
    $resp['avg_cost_ms'] = 0;
    $resp["set_ok"] = 0;
    $resp["get_ok"] = 0;

    $d = 0;
    generate($servers, $key, $value, $resp, $records, $d);
    // file
    writeSummary($resp, $size, 1);
    writeStatus($records);
}


function generate(&$servers, &$key, &$value, &$resp, &$records, &$d) {
    do {
        $cost = 0;
        $start = time();
        run($servers, $key, $value, $resp, $cost);
        $d += time() - $start;
        $resp["query"] = $resp["query"] + 2;
        $resp['cost_ms/10'] += $cost;
        $resp['avg_cost_ms'] = ($resp['cost_ms/10'] / 10.0) / $resp["query"];
        // status
        $k = $cost . "";  // =>  $count
        if (isset($records[$k])) {
            $records[$k] = $records[$k] + 2;
        } else {
            $records[$k] = 2;
        }
    } while ($d <= 3800);
}


function run($servers, &$key, &$value, &$resp, &$cost) {
    $redis = new Redis();
    foreach ($servers as $k => $v) {
        $res = $redis->pconnect($v['host'], $v['port'], 1);
        if (!$res) {
            continue;
        }
        try {
            // set
            $start = microtime(true);
            $r = $redis->set($key, $value, 3700);
            $cost += cost(microtime(true), $start);
            if ($r) {
                $resp["set_ok"] += 1;
            }

            // get
            $start = microtime(true);
            $r = $redis->get($key);
            $cost += cost(microtime(true), $start);
            if (strcmp($r, $value) == 0) {
                $resp["get_ok"] += 1;
            }
        } catch (Exception $e) {
            $f = "/tmp/error.log";
            $content = var_export($e) . "\n";
            file_put_contents($f, $content, FILE_APPEND);
        }
        if ($res) {
            $redis->close();
        }
    }
    $redis = null;
}


function writeSummary(&$resp, $size, $docker) {
    if ($docker == 1) {
        $f = "/tmp/docker/" . $size . "_KB_docker_status.txt";

    } else if ($docker == 0) {
        $f = "/tmp/docker/" . $size . "_KB_non-docker_status.txt";
    }
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

function writeStatus(&$records) {
    uksort($records, "byCost");
    foreach ($records as $cost => $count) {
        echo $cost . "," . $count . "\n";
    }
}

function byCost($k1, $k2) {
    return doubleval($k1) - doubleval($k2) > 0;
}