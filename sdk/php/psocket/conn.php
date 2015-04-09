<?php
/**
 * Created by PhpStorm.
 * @author: zhaoxi(LinZuxiong)
 * Date: 15-02-04
 * Time: 14:50
 */
require_once __DIR__ . '/../tests/format.php';

/**
 * Class (Conn)
 * @author: zhaoxi
 * @package psocket
 */
class Conn {
    private $redis;

    public function __construct() {
        try {
            $this->redis = new Redis();
        } catch (Exception $e) {
            echo "Exception..." . "\n";
            var_dump($e);
            echo "\n";
        }
    }

    public function __destruct() {
        $this->redis = null;
    }

    public function write($data) {
        $res = $this->redis->pconnect("/var/run/shadow_agent.sock");
        if (!$res) {
            return;
        }
        try {
            echo "connection status:";
            echo var_export($res) . "\n\n";
            $key = "PHP_key_testing";
//        $key = $data;
            $value = "PHP_value_testing";
            $value = $data;
            $res = "";
            $pid = getmypid();
            $fake = array();
            $fake['AppName'] = "CPC";
            $fake['Key'] = $key;
            $fake['cmd'] = "set";
            $fake['Async'] = true;
            $fake['PID'] = $pid;
            //-------------------msg pack
            $start = microtime(true);
            $compress = msgpack_pack($fake);
            $res = $this->redis->info($compress);
            $end = microtime(true);
            displayCost($end, $start);
            echo "FakeRedis result:\n";
            var_dump($res);
            //-------------------msg pack


            //-------------------info test
//            $_38bs = "12345678901234567890123456789012345678";
//            $res = $this->redis->info($_38bs);
            //-------------------info test


            //-------------------real redis command
            $t = 5 * 60;
            $start = microtime(true);
            $s = 8192;
            $key = buildKey(10);
            $value = build($s);
            $complex['age'] = $s;
            $complex['v'] = $value;
            $complex['simple'] = true;
            $msg = msgpack_pack($complex);
            $res = $this->redis->setex($key, 300, $msg);
//            $res = $this->redis->set($key, $value, $t);
            $end = microtime(true);
            displayCost($end, $start);
            echo "set result:\n";
            var_dump($res);
            //-------------------real redis command


            //-------------------real redis command
//            $start = microtime(true);
//            $res = $this->redis->get($key);
//            $end = microtime(true);
//            displayCost($end, $start);
//            echo "get result:\n";
//            var_dump($res);
            //-------------------real redis command


            //----------------------------------------------
            echo "\nClose ok.\n";
        } catch (Exception $e) {
            echo "Exception..." . var_export($e) . "\n";
        }
        $this->redis->close();
        echo "==================== PHP write done.================" . "\n";
    }
}
