<?php
/**
 * Created by PhpStorm.
 * User: LinZuxiong
 * Date: 15-02-27
 * Time: 14:46
 */


function buildKey($bytes) {
    return "_StressTesting_" . getmypid() . "_" . build($bytes);
}

// unit : KB
function buildValue($kbs) {
    $kbs = $kbs * 1024;
    return build($kbs);
}


/**
 * @param $bytes
 * @return string
 */
function build($bytes) {
    $characters = '0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz';
    $result = '';
    for ($i = 0; $i < $bytes; $i++) {
        $result .= $characters[mt_rand(0, 61)];
    }
    return $result;
}

/**
 * zoom in: 1000*10
 * @param $time
 * @return int
 */
function formatMicro($time) {
    $v = $time * 1000 * 10;
    return intval($v);
}


/**
 * @param $end
 * @param $start
 * @return int   N(/10ms) == ?ms
 */
function cost($end, $start) {
    return formatMicro($end) - formatMicro($start);
}

function displayCost($end, $start) {
    $cost = formatMicro($end) - formatMicro($start);
    echo "== " . ($cost) . "(/10ms) === ms :  " . ($cost / 10) . "\n";
}