<?php
include('conn.php');
require_once __DIR__ . '/../tests/format.php';
/**
 * Created by PhpStorm.
 * @author: zhaoxi(LinZuxiong)
 * Date: 15-02-04
 * Time: 16:46
 */


$conn = new Conn();
$data = buildValue(1);


$times = 1;
for ($i = 0; $i < $times; $i++) {
    $conn->write($data);
}

?>
