<?php

/**
 * Created by PhpStorm.
 * User: zhaoxi (zhaoxi@mogujie.com)
 * Date: 15-02-27
 * Time: 16:21
 */
require __DIR__ . '/factory.php';

// ttl :  Time To Live

class Shadow_Cache implements Shadow_Factory {

    const MAX_VALUE_BYTES = 131072; // 128 KiB
//    const MAX_VALUE_BYTES = 262144; // 256 KiB
//    const MAX_VALUE_BYTES = 10485760; // 10 MiB


    /**
     * @var Shadow_Cache
     */
    private static $_instance = null;
    private static $_lastAppName = null;

    /**
     * @param $app
     * @return Shadow_Cache
     * @throws Exception
     */
    public static function build($app) {
        if (!is_string($app)) {
            throw new Exception("Please input AppName that is string!");
        }
        if ($app == null || strlen($app) < 1) {
            throw new RuntimeException("Parameter is empty!Please check it!");
        }
        if (self::$_instance != null) {
            if (strcmp(self::$_lastAppName, $app) == 0) {
                return self::$_instance;
            } else {
                // destroy the old one!
                self::$_instance->close();
                self::$_instance = null;
            }
        }
        // Prepare to new one...
        self::$_lastAppName = $app;
        return new self();
    }

    /**
     * @var Redis
     */
    private $_redis = null;
    private $_socket = "/var/run/shadow/agent.sock";

    private function __construct() {
        $this->_redis = new Redis();
    }

    public function __destruct() {
        self::$_instance = null;
        $this->close();
        if ($this->_redis !== null) {
            $this->_redis = null;
        }
    }

    private function close() {
        try {
            if ($this->_redis !== null) {
                $this->_redis->close();
            }
        } catch (Exception $e) {
        }
        $this->_redis = null;
    }
    //===================================================================================
    /**
     * @param $ttl , the unit is second
     * @return int
     */
    public function maxTTL($ttl) {
        if ($ttl === null || $ttl >= 604800 || !is_numeric($ttl)) {
            // 7*24*3600
            return 604800;
        }
        return intval($ttl);
    }

    /**
     * UTF-8 bytes: length <= 1024
     * @param $key
     * @return string
     */
    public function handleKey($key) {
        if ($key == null || $key == null) {
            throw new RuntimeException('Param is error, the key: ' . $key);
        }
        if (mb_strlen($key, "UTF-8") > 1024) {
            $k = sha1($key, false);
            if ($k) {
                return $k;
            }
            throw new RuntimeException("can not sha1: " . $key);
        }
        return $key;
    }

    public function sendFake($key, $async, $cmd) {
        $fake = array();
        $fake['AppName'] = self::$_lastAppName;
        $fake['Key'] = $key;
        $fake['Cmd'] = $cmd;
        $fake['Async'] = $async;
        $compress = msgpack_pack($fake);
        $this->_redis->info($compress);
        return;
    }
    //===================================================================================


    //======================= commands#string =================================
    /**
     * @param $key string
     * @param bool $async
     * @return bool
     * @throws Exception
     */
    public function delete($key, $async = false) {
        $key = $this->handleKey($key);
        $res = $this->_redis->pconnect($this->_socket, 0, 1);
        if (!$res) {
            throw new Exception("can not connect the UnixSocket");
        }
        try {
            //--------------------------------fake-----------------------------
            $this->sendFake($key, $async, "DELETE");
            //--------------------------------fake-----------------------------
            $this->_redis->delete($key);
            $this->_redis->close();
            return true;
        } catch (Exception $e) {
            if ($this->_redis !== null) {
                $this->_redis->close();
            }
            throw $e;
        }
    }

    /**
     * 即将废弃的API
     * @param $keys array
     * @return array
     */
    public function deleteMulti($keys) {
        return $this->deleteBatch($keys, true);
    }

    /**
     * @param $keys array
     * @param bool $async
     * @return array
     */
    public function deleteBatch($keys, $async = false) {
        if (!is_array($keys)) {
            throw new RuntimeException("Illegal Param! k_2_v is not array.");
        }
        if (count($keys) < 1) {
            return true;
        }
        $result = array();
        // TODO batch
        foreach ($keys as $key) {
            try {
                $this->delete($key, $async);
                $result[$key] = true;
            } catch (Exception $e) {
                $result[$key] = false;
            }
        }
        return $result;
    }

    /**
     * 对某一个key值进行超时限制
     * @param $key string
     * @param $timeout int
     * @param bool $async
     * @return bool
     * @throws Exception
     */
    public function expire($key, $timeout, $async = false) {
        $key = $this->handleKey($key);
        $res = $this->_redis->pconnect($this->_socket, 0, 1);
        if (!$res) {
            throw new Exception("can not connect the UnixSocket");
        }
        try {
            //--------------------------------fake-----------------------------
            $this->sendFake($key, $async, "EXPIRE");
            //--------------------------------fake-----------------------------
            $r = $this->_redis->expire($key, $timeout);
            $this->_redis->close();
            return $r;
        } catch (Exception $e) {
            if ($this->_redis !== null) {
                $this->_redis->close();
            }
            throw $e;
        }
    }


    /**
     * @param $key
     * @param $ttl
     * @param $value
     * @param bool $async
     * @return bool
     * @throws Exception
     */
    public function setex($key, $ttl, $value, $async = false) {
        $key = $this->handleKey($key);
        $ttl = $this->maxTTL($ttl);
        $res = $this->_redis->pconnect($this->_socket, 0, 1);
        if (!$res) {
            throw new Exception("can not connect the UnixSocket");
        }
        try {
            $msg = msgpack_pack($value);
            if (mb_strlen($msg, "UTF-8") > self::MAX_VALUE_BYTES) {
                $this->close();
                throw new RuntimeException("Illegal value length. It beyonds " . self::MAX_VALUE_BYTES . " bytes.");
            }
            //--------------------------------fake-----------------------------
            $this->sendFake($key, $async, "SETEX");
            //--------------------------------fake-----------------------------
            $r = $this->_redis->setex($key, $ttl, $msg);
            $this->_redis->close();
            return $r;
        } catch (Exception $e) {
            if ($this->_redis !== null) {
                $this->_redis->close();
            }
            throw $e;
        }
    }

    /**
     * 这是兼容API,老的主站代码需要兼容.即将废弃的API
     * @param $key string
     * @param $value string | array()
     * @param $ttl int
     * @param $compressed bool    Unnecessary Param!@zhaoxi comments
     * @param bool $async
     * @throws Exception
     * @return bool
     */
    public function set($key, $value, $ttl = null, $compressed = true, $async = false) {
        return $this->setex($key, $ttl, $value, $async);
    }

    /**
     * @param $key string
     * @return array|bool|string
     * @throws Exception
     */
    public function get($key) {
        $key = $this->handleKey($key);
        $res = $this->_redis->pconnect($this->_socket, 0, 1);
        if (!$res) {
            throw new Exception("can not connect the UnixSocket");
        }
        try {
            //--------------------------------fake-----------------------------
            $this->sendFake($key, false, "GET");
            //--------------------------------fake-----------------------------
            $r = $this->_redis->get($key);
            $r = msgpack_unpack($r);
            $this->_redis->close();
            return $r;
        } catch (Exception $e) {
            if ($this->_redis !== null) {
                $this->_redis->close();
            }
            throw $e;
        }
    }

    /**
     * 这是兼容API,老的主站代码需要兼容.即将废弃的API
     * @param $keys array 查询条件
     * @param $compressed bool    Unnecessary Param!@zhaoxi comments 默认是开启的
     * @return array | bool
     */
    public function getMulti($keys, $compressed = true) {
        return $this->getBatch($keys);
    }

    /**
     * 进行批量查询
     * @param $keys array
     * @return array | bool
     */
    public function getBatch($keys) {
        if (!is_array($keys)) {
            throw new RuntimeException("Illegal Param! keys is not array.");
        }
        if (count($keys) < 1) {
            return true;
        }
        $res = array();
        if ($keys == null || count($keys) < 1) {
            return $res;
        }
        foreach ($keys as $key) {
            $res[$key] = $this->get($key);
        }
        return $res;
    }

    /**
     * 这是兼容API,老的主站代码需要兼容.即将废弃的API
     * @param $k_2_v
     * @param $ttl int
     * @param $compressed bool    Unnecessary Param!@zhaoxi comments
     * @return array
     */
    public function setMulti($k_2_v, $ttl = null, $compressed = true, $async = false) {
        return $this->setBatch($k_2_v, $ttl, $async);
    }

    /**
     * @param $k_2_v
     * @param $ttl
     * @param bool $async
     * @return array
     */
    public function setBatch($k_2_v, $ttl, $async = false) {
        if (!is_array($k_2_v)) {
            throw new RuntimeException("Illegal Param! k_2_v is not array.");
        }
        $res = array();
        if (count($k_2_v) < 1) {
            return $res;
        }
        // TODO batch
        foreach ($k_2_v as $key => $value) {
            try {
                $res[$key] = $this->setex($key, $ttl, $value, $async);
            } catch (Exception $e) {
                $res[$key] = false;
            }
        }
        return $res;
    }
    //======================= commands#string  end ============================


    //======================= commands#list ===================================
    public function lrem($key, $count, $value, $async = false) {
    }

    public function left() {
    }

    public function ltrim() {
    }

    public function lrange() {
    }

    public function lrangeMulti() {
    }

    public function push() {
    }

    public function pushMulti() {
    }

    public function pop() {
    }

    public function popMulti() {
    }
    //======================= commands#list end ===============================


    //======================= commands#hash ===================================
    public function hset() {
    }

    public function hmset() {
    }

    public function hget() {
    }

    public function hmget() {
    }

    public function hdel() {
    }

    public function hgetall() {
    }

    public function hlen() {
    }
    //======================= commands#hash end ===============================


    //======================= commands#set ====================================
    public function sadd() {
    }

    public function spop() {
    }

    public function srem() {
    }

    public function smembers() {
    }

    public function scard() {
    }
    //======================= commands#set end ================================


    //======================= commands#sorted_set =============================
    /**
     * sorted_set
     */
    public function zadd() {
    }

    /**
     * sorted_set
     */
    public function zscore() {
    }

    /**
     * sorted_set
     */
    public function zrank() {
    }

    /**
     * sorted_set
     */
    public function zrevrank() {
    }

    /**
     * sorted_set
     */
    public function zrem() {
    }

    /**
     * sorted_set
     */
    public function zremrangebyrank() {
    }

    /**
     * sorted_set
     */
    public function zremrangebyscore() {
    }

    /**
     * sorted_set
     */
    public function zrange() {
    }

    /**
     * sorted_set
     *
     */
    public function zrangebyscore() {
    }

    /**
     * sorted_set
     */
    public function zrevrangebyscore() {
    }

    /**
     * sorted_set
     */
    public function zrevrange() {
    }

    /**
     * sorted_set
     */
    public function zcount() {
    }
    //======================= commands#sorted_set end =========================
}
