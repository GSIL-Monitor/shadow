<?php

/**
 * TODO
 * @author ziyuan
 * @date January 04, 2015
 * @project shadow
 * @description cache instance/SDK
 */
include 'conn.php';
include 'parser.php';

class ShadowCache {

    const DEFAULT_ENCODING = "UTF-8";
    const CRLF = "\r\n";

    // method name, with redis
    const GET = "$3\r\nGET\r\n";
    const SET = "$3\r\nSET\r\n";
    const SETEX = "$5\r\nSETEX\r\n";
    const MGET = "$4\r\nMGET\r\n";
    const MSET = "$4\r\nMSET\r\n";
    const DEL = "$3\r\nDEL\r\n";
    const EXPIRE = "$6\r\nEXPIRE\r\n";
    const MULTI = "$5\r\nMULTI\r\n";
    const EXEC = "$4\r\nEXEC\r\n";
    const LREM = "$4\r\nLREM\r\n";
    const LLEN = "$4\r\nLLEN\r\n";
    const LTRIM = "$5\r\nLTRIM\r\n";
    const LRANGE = "$6\r\nLRANGE\r\n";
    const LPUSH = "$5\r\nLPUSH\r\n";
    const RPOP = "$4\r\nRPOP\r\n";
    const HSET = "$4\r\nHSET\r\n";
    const HMSET = "$5\r\nHMSET\r\n";
    const HGET = "$4\r\nHGET\r\n";
    const HMGET = "$5\r\nHMGET\r\n";
    const HDEL = "$4\r\nHDEL\r\n";
    const HINCRBY = "$7\r\nHINCRBY\r\n";
    const HGETALL = "$7\r\nHGETALL\r\n";
    const HLEN = "$4\r\nHLEN\r\n";
    const SADD = "$4\r\nSADD\r\n";
    const SPOP = "$4\r\nSPOP\r\n";
    const SREM = "$4\r\nSREM\r\n";
    const SMEMBERS = "$8\r\nSMEMBERS\r\n";
    const SCARD = "$5\r\nSCARD\r\n";
    const ZADD = "$4\r\nZADD\r\n";
    const ZREM = "$4\r\nZREM\r\n";
    const ZREVRANGE = "$9\r\nZREVRANGE\r\n";
    const ZREVRANGEBYSCORE = "$16\r\nZREVRANGEBYSCORE\r\n";
    const ZREMRANGEBYSCORE = "$16\r\nZREMRANGEBYSCORE\r\n";
    const ZRANGEBYSCORE = "$13\r\nZRANGEBYSCORE\r\n";
    const ZCOUNT = "$6\r\nZCOUNT\r\n";
    const ZREVRANK = "$8\r\nZREVRANK\r\n";
    const ZRANGE = "$6\r\nZRANGE\r\n";
    const ZSCORE = "$6\r\nZSCORE\r\n";
    const ZRANK = "$5\r\nZRANK\r\n";
    const ZREMRANGEBYRANK = "$15\r\nZREMRANGEBYRANK\r\n";

    //TODO 这个变量干吗吗? 注释和后面的代码表达不一样?
    private $_cache = null; // cache 唯一标识符

    /**
     * @var null|Conn
     */
    private $_conn = null; // 本地连接

    /**
     * @var null|ShadowCache
     */
    private static $_instance = null;
    private static $_lastAppName = null;


    /**
     * Simple Factory.<br/>
     * 静态方法，获得实例.
     * @param $appname
     * @return null|ShadowCache
     * @throws RuntimeException
     */
    public static function build($appname) {
        if ($appname == null || strlen($appname) < 1) {
            throw new RuntimeException("Parameter is empty!Please check it!");
        }
        if (self::$_instance != null) {
            if (strcmp(self::$_lastAppName, $appname) == 0) {
                return self::$_instance;
            } else {
                // destroy the old one!
                self::$_instance->close();
                self::$_instance = null;
            }
        }
        // Prepare to new one...
        self::$_lastAppName = $appname;
        return new self($appname);
    }

    /**
     * 构造函数
     * 获取Redis的配置和实例情况
     * @param cache string,即相关cache的标识
     * @throws RuntimeException
     */
    private function __construct($appname) {
        $this->_cache = array('appname' => $appname);
        $this->_conn = Conn::build();
        if (false === $this->_conn) {
            // log  can not get one connection for connecting Agent.
            self::$_instance = null;
            self::$_lastAppName = null;
            $this->_cache = null;
            $this->_conn = null;
            throw new RuntimeException("Not available instance!Please check it!");
        }
    }

    /**
     * 析构函数，暂时没有作用
     */
    public function __destruct() {
        $this->close();
    }

    /**
     * 显式的关闭该次实例
     */
    public function close() {
        $this->_cache = null;
        if ($this->_conn != null) {
            $this->_conn->close();
            // finally
            $this->_conn = null;
        }
    }

    /**
     * 获得连接本地agent的实例
     */
    public function connect() {
    }

    //==========================================================
    //====================== String Commands ===================
    /**
     * 删除一个key
     * @param $key string
     * @param bool $asyn
     * @return bool
     */
    public function delete($key, $asyn = false) {
        $len = count($this->_cache) + 1; // 这边加上新的数据
        $cmd = self::DEL;
        $params[] = $key;
        $res = $this->doCmd($cmd, $params, $asyn);
        return $res;
    }

    /**
     * 删除多个key值
     * @param $keys array
     * @param bool $asyn
     * @return array
     */
    public function deleteMulti(array $keys, $asyn = false) {
        // Agent端无法保证连接的唯一性，这边不能使用multi命令
        $res = true;
        foreach ($keys as $key) {
            $res = $res && (bool)self::delete($key, $asyn);
        }
        return (bool)$res;
    }

    /**
     * 对某一个key值进行超时限制
     * @Param $key string
     * @param $timeout int
     */
    public function expire($key, $timeout, $asyn = false) {
        if (!$key || is_array($key)) {
            throw new Exception('Param is error, ' . $key);
        }
        $cmd = self::EXPIRE;
        $params[] = $key;
        $params[] = $timeout;
        $res = $this->doCmd($cmd, $params, $asyn);
        return $res;
    }

    /**
     * 设置缓存，这边如果业务方传入的值为array，则使用msgpack进行压缩,并且加上压缩标志位
     * @param $key string
     * @param $value string | array()
     * @param $expiration int
     * @param $encode bool
     * @param bool $asyn
     * @throws Exception
     * @return bool
     */
    public function set($key, $value, $expiration = null, $encode = true, $asyn = false) {
        if ($key == null || $key == null) {
            return false;
        }

        if ($expiration == null) {
            $cmd = self::SET;
            $params[] = $key;
        } else {
            $cmd = self::SETEX;
            $params[] = $key;
            $params[] = $expiration;
        }
        if ($encode || is_array($value)) {
            $value = msgpack_pack($value);
            if (mb_strlen($value) > 50 * 1024) {
                // 这边是一个安全限制
                throw new Exception("Your package is too big to send");
            }
        }
        $params[] = $value;
        $res = $this->doCmd($cmd, $params, $asyn);
        return $res;
    }

    /**
     * 从缓存中获得数据
     * @param $key string
     * @param $encode bool 是否压缩，默认开启
     * @return string | array() | bool
     */
    public function get($key, $encode = true) {
        if ($key == null) {
            return false;
        }
        $cmd = self::GET;
        $params[] = $key;
        $res = $this->doCmd($cmd, $params);
        if ($encode) {
            $res = msgpack_unpack($res);
        }
        return $res;
    }

    /**
     * 进行批量查询
     * @param $keys array 查询条件
     * @param $encode bool 默认是开启的
     * @return array | bool
     */
    public function getMulti(array $keys, $encode = true) {
        return self::getBatch($keys, $encode);
    }

    /**
     * 进行批量查询
     * @param $keys array
     * @param $encode bool
     * @return array | bool
     */
    public function getBatch(array $keys, $encode = true) {
        foreach ($keys as $key) {
            $res[$key] = self::get($key, $encode);
        }
        return $res;
    }

    /**
     * set 多个值
     * @param $values array(key, value)
     * @param $expiration int
     * @param $encode bool
     * @return array
     */
    public function setMulti(array $values, $expiration = null, $encode = true, $asyn = false) {
        $res = true;
        foreach ($values as $key => $value) {
            $res = $res && (bool)self::set($key, $value, $expiration, $encode, $asyn);
        }
        return $res;
    }
    //======================= String end ===============================
    //==================================================================

    //===================================================================
    //====================== List Command ===============================
    //默认队列名
    public static $queueName = 'mq';

    //改变队列
    public function changeQueuePool($queueName = 'mq') {
        self::$queueName = $queueName;
        return $this;
    }

    /**
     * 从list中移除元素
     * @param $value string
     * @param $count int
     * @param $queueName string
     * @return int
     */
    public function lrem($value, $count = 0, $queueName, $asyn = false) {
        $queueName = is_null($queueName) ? self::$queueName : $queueName;
        $res = $this->doCmd(self::LREM, array($queueName, $value, $count), $asyn);
        return $res;
    }

    /**
     * 当前有多少剩余
     * @param $queueName string
     * @return int
     */
    public function left($queueName = null) {
        $queueName = is_null($queueName) ? self::$queueName : $queueName;
        $res = $this->doCmd(self::LLEN, array($queueName));
        return $res;
    }

    /**
     * Trim List
     * @param $start int
     * @param $end int
     * @param $queueName string
     * @return int
     */
    public function ltrim($start, $end, $queueName = null) {
        $queueName = is_null($queueName) ? self::$queueName : $queueName;
        $res = $this->doCmd(self::LTRIM, array($queueName, $start, $end));
        return $res;
    }


    /**
     * 单纯获取队列中的元素值，但不出队
     * @param $start int
     * @param $end int
     * @param $queueName string
     * @param $encode bool
     * @return array | bool
     */
    public function lrange($start = 0, $end = -1, $queueName = null, $encode = true) {
        $queueName = is_null($queueName) ? self::$queueName : $queueName;
        $res = $this->doCmd(self::LRANGE, array($queueName, $start, $end));

        if (empty($res)) return null;
        if (!$encode) {
            return is_null($res) ? false : $res;
        }

        $return = array();
        foreach ($res as $k => $v) {
            unset($res[$k]);

            if (false === $v) {
                $v = null;
            }
            if (is_null($v)) {
                break;
            }

            if (true == $encode) {
                $temp = msgpack_unpack($v);
                $return[] = $temp;
            } else {
                $return[] = $v;
            }
        }

        return $return;
    }

    /**
     * Retrieve multiple items
     * @param $queueName_arr array
     * @param $offset int
     * @param $end int
     * @param $prefix string
     * @param $encode bool
     * @return array | bool
     */
    public function lrangeMulti(array $queueName_arr, $offset = 0, $end = -1, $prefix = '', $encode = true) {
        if (0 == count($queueName_arr)) return false;

        $res = array();
        foreach ($queueName_arr as $queueName) {
            $res[$queueName] = $this->lrange($offset, $end, $prefix . $queueName, $encode);
        }
        return $res;

        // $i = 0;
        // $commands[$i]['cmd'] = self::MULTI;
        // $commands[$i++]['params'][] = $pool;
        // foreach ($queueName_arr as $queueName) {
        // $pool = $prefix.$queueName;
        // $commands[$i]['cmd'] = self::LRANGE;
        // $commands[$i]['params'][] = $pool;
        // $commands[$i]['params'][] = intval($offset);
        // $commands[$i++]['params'][] = intval($end);
        // }
        // $commands[$i]['cmd'] = self::exec;
        // $commands[$i]['params'][] = $pool;
        // $commands = $this->buildMulticmd($commands);
        // try{
        // $res = $this->_conn->pipeline($commands);

        // if(!is_array($res) || 0 == count($res)) return null;
        // $return = array();
        // foreach($queueName_arr as $k=>$v){
        // $return[$v] = $res[$k];
        // }
        // unset($res);
        // return $return;
        // } catch(Exception $e){
        // return false;
        // }
    }

    /**
     * 把变量入队, 不能是空变量;
     * @param $value string
     * @param $queueName string
     * @param $queueLength int
     * @param $encode bool
     * @param $asyn bool
     * @return bool
     */
    public function push($value, $queueName = null, $queueLength = 0, $encode = true, $asyn = false) {
        if (mb_strlen($value) > 50 * 1024) {
            // 这边是一个安全限制
            throw new Exception("Your package is too big");
        }
        if (is_null($value)) return null;

        $queueName = is_null($queueName) ? self::$queueName : $queueName;
        if (true == $encode) {
            $value = msgpack_pack($value);
        }
        $res = $this->doCmd(self::LPUSH, array($queueName, $value), $asyn);

        if (0 < $queueLength && $res > $queueLength) {
            $this->ltrim(0, $queueLength - 1, $queueName);
        }
        return $res;
    }

    /**
     * @param $values array
     * @param $queueName string
     * @param $encode bool
     * @param $asyn bool
     * @return bool
     */
    public function pushMulti(array $values, $queueName = null, $encode = true, $asyn = false) {
        $queueName = is_null($queueName) ? self::$queueName : $queueName;

        $res = true;
        foreach ($values as $value) {
            $res = $res && $this->push($value, $queueName, $encode, $asyn);
        }
        return $res;
        // $i = 0;
        // $commands[$i]['cmd'] = self::MULTI;
        // $commands[$i++]['params'][] = $queueName;
        // foreach ($values as $value) {
        // $commands[$i]['cmd'] = self::LPUSH;
        // $commands[$i++]['params'][] = $queueName;
        // if ($encode) {
        // $value = msgpack_pack($value);
        // }
        // $commands[$i++]['params'][] = $value;
        // }
        // $commands[$i]['cmd'] = self::EXEC;
        // $commands[$i]['params'][] = $queueName;
        // $commands = $this->buildMulticmd($commands);
        // $res = $this->_conn->pipeline($commands);

        // return @$res[0];
    }

    /**
     * 出队,单个
     * @param $queueName string
     * @param $encode bool
     * @return string
     */
    public function pop($queueName = null, $encode = true) {
        $queueName = is_null($queueName) ? self::$queueName : $queueName;

        $res = $this->doCmd(self::RPOP, array($queueName));

        if (false === $res) return null;
        if (true == $encode) {
            $temp = msgpack_unpack($res);
            $res = is_null($temp) ? null : $temp;
        }
        return $res;
    }

    /**
     * 出队列，获取队列中多个值
     * @param $length int
     * @param $queueName string
     * @param $encode bool
     * @return array | bool
     */
    public function popMulti($length = 5, $queueName = null, $encode = true) {
        $queueName = is_null($queueName) ? self::$queueName : $queueName;

        $res = array();
        for ($i = 0; $i < $length; $i++) {
            $res[] = $this->pop($queueName, $encode);
        }
        return $res;
        // $i = 0;
        // $commands[$i]['cmd'] = self::MULTI;
        // $commands[$i++]['params'][] = $queueName;
        // for($j = 0; $j < $length; $j++){
        // $commands[$i]['cmd'] = self::RPOP;
        // $commands[$i++]['params'][] = $queueName;
        // }
        // $commands[$i]['cmd'] = self::EXEC;
        // $commands[$i]['params'][] = $queueName;
        // $commands = $this->buildMulticmd($commands);
        // $res = $this->_conn->pipeline($commands);
        // $return = array();
        // if(!empty($res)){
        // foreach($res as $k => $v){
        // unset($res[$k]);

        // if(false === $v){
        // $v = null;
        // }
        // if(is_null($v)) break;

        // if(true == $encode){
        // $temp  = msgpack_unpack($v);
        // if(0<$this->gzcompress_level && !is_null($res)){
        // if(!$temp){
        // $res = gzuncompress($v);
        // $temp = msgpack_unpack($v);
        // }
        // }
        // $v = $temp;
        // }
        // $return[] = $v;
        // }
        // }
        // return empty($return) ? null : $return;
    }
    //====================== List End ===================================
    //===================================================================

    //====================== Hash Start =================================
    //===================================================================
    /**
     * redis hash
     */
    public static $hashPool = 'mghash';

    /**
     * 改变hash
     */
    public function changeHash($hashPool = 'mghash') {
        self::$hashPool = $hashPool;
        return $this;
    }

    /**
     * redis hash hset
     * @param $key string
     * @param $value string | array
     * @param null $hashPool
     * @param null $expiration
     * @param bool $asyn
     * @throws Exception
     * @return user define
     * 设置单个Key值
     */
    public function hset($key, $value, $hashPool = null, $expiration = null, $asyn = false) {
        if (mb_strlen($value) > 50 * 1024) {
            // 这边是一个安全限制
            throw new Exception("Your package is too big");
        }
        if (is_null($key) || is_null($value)) {
            return false;
        }

        $hashPool = is_null($hashPool) ? self::$hashPool : $hashPool;

        $res = intval($this->doCmd(self::HSET, array($hashPool, $key, $value), $asyn));

        if (!is_null($expiration)) {
            $expiration = $this->getLifetime($expiration);
            $this->doCmd(self::EXPIRE, array($hashPool, $expiration), $asyn);
        }
        return $res;
    }

    /*
    * redis hash hmset
    * @param $item array key=>value
    * 设置多个key值
    */
    public function hmset(array $items, $hashPool = null, $expiration = null, $asyn = false) {
        $lifetime = $this->getLifetime($expiration);

        $hashPool = is_null($hashPool) ? self::$hashPool : $hashPool;

        $params[] = $hashPool;
        foreach ($items as $key => $value) {
            $params[] = $key;
            $params[] = $value;
        }

        $res = intval($this->doCmd(self::HMSET, $params), $asyn);

        if (!is_null($expiration)) {
            $expiration = $this->getLifetime($expiration);
            $this->doCmd(self::EXPIRE, array($hashPool, $expiration), $asyn);
        }

        return $res;
    }

    /**
     * redis hash get
     * @param $key string
     * @param null $hashPool
     * @return array
     */
    public function hget($key, $hashPool = null) {
        if (is_null($key)) {
            return null;
        }

        $hashPool = is_null($hashPool) ? self::$hashPool : $hashPool;
        $res = $this->doCmd(self::HGET, array($hashPool, $key));
        return ($res === false) ? null : $res;
    }

    /**
     * redis hash hmget
     * @获取key对应的value，如果不存在，返回null
     */
    public function hmget($keys, $hashPool = null) {
        if (is_null($keys)) {
            return null;
        }

        $hashPool = is_null($hashPool) ? self::$hashPool : $hashPool;
        $params[] = $hashPool;
        foreach ($keys as $key) {
            $params[] = $key;
        }
        $res = $this->doCmd(self::HMGET, $params);
        if (!is_array($res)) {
            return false;
        }
        foreach ($keys as $key) {
            $ret[$key] = array_shift($res);
        }
        foreach ($ret as $k => $v) {
            if ($v === false) $res[$k] = null;
        }
        return $ret;
    }

    /**
     * redis hash hdel
     * 删除key
     */
    public function hdel($key, $hashPool = null, $asyn = false) {
        if (is_null($key)) {
            return true;
        }

        $hashPool = is_null($hashPool) ? self::$hashPool : $hashPool;
        $res = $this->doCmd(self::HDEL, array($hashPool, $key), $asyn);
        return $res;
    }

    /**
     * redis hash hgetall
     */
    public function hgetall($hashPool = null) {
        $hashPool = is_null($hashPool) ? self::$hashPool : $hashPool;
        $res = $this->doCmd(self::HGETALL, array($hashPool));
        return $res;
    }

    /**
     * redis hash hlen
     */
    public function hlen($hashPool = null) {
        $hashPool = is_null($hashPool) ? self::$hashPool : $hashPool;
        $res = $this->doCmd(self::HLEN, array($hashPool));
        return $res;
    }
    //======================= Hash End ==========================
    //===========================================================

    //======================= Sorted Set ========================
    //===========================================================
    /**
     * sorted set 默认key
     */
    public static $sortedSet = 'redis_sortedset';

    //改变set
    public function changeSortedSet($sortedSet = 'redis_sortedset') {
        self::$sortedSet = $sortedSet;
        return $this;
    }

    /**
     * redis sorted set
     * zadd
     */
    public function zadd($keys, $values, $sortedSet = null, $asyn = false) {
        if (is_null($keys) || is_null($values)) return false;

        $sortedSet = is_null($sortedSet) ? self::$sortedSet : $sortedSet;
        $args = array();
        if (is_array($keys) && is_array($values)) {
            if (count($keys) == 0 || count($values) == 0) return false;
            if (count($keys) == count($values)) {
                for ($i = 0; $i < count($keys); $i++) {
                    array_push($args, $values[$i]);
                    array_push($args, $keys[$i]);
                }
            }
        } else if (!is_array($keys) && !is_array($values)) {
            array_push($args, $values);
            array_push($args, $keys);
        } else {
            return false;
        }
        array_unshift($args, $sortedSet);
        $res = $this->doCmd(self::ZADD, $args, $asyn);

        return $res;
    }

    /**
     * redis sorted set
     * zscore
     */
    public function zscore($key, $sortedSet = null) {
        if (is_null($key)) return null;
        $sortedSet = is_null($sortedSet) ? self::$sortedSet : $sortedSet;

        $res = $this->doCmd(self::ZSCORE, array($sortedSet, $key));
        if (false === $res) $res = null;
        return $res;
    }

    /**
     * redis sorted set
     * zscore
     */
    public function zrank($key, $sortedSet = null) {
        if (is_null($key)) return null;
        $sortedSet = is_null($sortedSet) ? self::$sortedSet : $sortedSet;
        $res = $this->doCmd(self::ZRANK, array($sortedSet, $key));
        $res = (false === $res) ? null : $res;
        return $res;
    }

    /**
     * redis sorted set
     * zrevrank
     */
    public function zrevrank($key, $sortedSet = null) {
        if (is_null($key)) return null;
        $sortedSet = is_null($sortedSet) ? self::$sortedSet : $sortedSet;
        $res = $this->doCmd(self::ZREVRANK, array($sortedSet, $key));
        $res = (false === $res) ? null : $res;
        return $res;
    }

    /**
     * redis sorted set
     * ZREM
     */
    public function zrem($keys, $sortedSet = null, $asyn = false) {
        if (is_null($keys)) return true;
        $sortedSet = is_null($sortedSet) ? self::$sortedSet : $sortedSet;
        $args = array();
        if (is_array($keys)) {
            if (count($keys) == 0) return true;
            $args = $keys;
        } else {
            array_push($args, $keys);
        }
        array_unshift($args, $sortedSet);
        $res = $this->doCmd(self::ZREM, $args, $asyn);
        return $res;
    }

    /**
     * redis sorted_set mul_zrem
     * zremrangebyrank
     */
    public function zremrangebyrank($start, $end, $sortedSet = null, $asyn = false) {
        if (is_null($start) || is_null($end)) return null;
        $sortedSet = is_null($sortedSet) ? self::$sortedSet : $sortedSet;
        $res = $this->doCmd(self::ZREMRANGEBYRANK, array($sortedSet, $start, $end));
        return $res;
    }

    /**
     * redis sorted set zremrangebyscore
     */
    public function zremrangebyscore($min, $max, $sortedSet = null, $asyn = false) {
        $sortedSet = is_null($sortedSet) ? self::$sortedSet : $sortedSet;
        $res = $this->doCmd(self::ZREMRANGEBYSCORE, array($sortedSet, $min, $max));
        return $res;
    }

    /**
     * redis sorted_set zrange
     */
    public function zrange($start, $end, $sortedSet = null, $WITHSCORES = false) {
        if (is_null($start) || is_null($end)) return null;
        if ($sortedSet == null) {
            $sortedSet = self::$sortedSet;
        }
        if ($WITHSCORES) {
            $res = $this->doCmd(self::ZRANGE, array($sortedSet, $start, $end, 'WITHSCORES'));
        } else {
            $res = $this->doCmd(self::ZRANGE, array($sortedSet, $start, $end));
        }
        return $res;
    }

    /**
     *    redis sorted set
     *    zrangebyscore
     */
    public function zrangebyscore($min_score, $max_score, $sortedSet = null, $offset = null, $count = null, $withscores = false) {
        if (is_null($min_score) || is_null($max_score)) return null;
        $sortedSet = is_null($sortedSet) ? self::$sortedSet : $sortedSet;
        $params[] = $sortedSet;
        $params[] = $min_score;
        $params[] = $max_score;
        if (!is_null($withscores) && $withscores) {
            $params[] = "WITHSCORES";
        }
        if (!is_null($offset) && !is_null($count)) {
            $params[] = "LIMIT";
            $params[] = $offset;
            $params[] = $count;
        }
        $res = $this->doCmd(self::ZRANGEBYSCORE, $params);
        return $res;
    }


    /**
     *    zrevrangebyscore
     */
    public function zrevrangebyscore($max_score, $min_score, $sortedSet = null, $offset = null, $count = null, $withscores = false) {
        if (is_null($max_score) || is_null($min_score)) return null;
        $sortedSet = is_null($sortedSet) ? self::$sortedSet : $sortedSet;
        $params[] = $sortedSet;
        $params[] = $min_score;
        $params[] = $max_score;
        if (!is_null($withscores) && $withscores) {
            $params[] = "WITHSCORES";
        }
        if (!is_null($offset) && !is_null($count)) {
            $params[] = "LIMIT";
            $params[] = $offset;
            $params[] = $count;
        }
        $res = $this->doCmd(self::ZREVRANGEBYSCORE, $params);
        return $res;
    }

    /**
     * redis sorted_set
     * zrevrange
     */
    public function zrevrange($start, $end, $sortedSet = null, $WITHSCORES = false) {
        if (is_null($start) || is_null($end)) return null;
        $sortedSet = is_null($sortedSet) ? self::$sortedSet : $sortedSet;

        if ($WITHSCORES) {
            $res = $this->doCmd(self::ZREVRANGE, array($sortedSet, $start, $end, 'WITHSCORES'));
        } else {
            $res = $this->doCmd(self::ZREVRANGE, array($sortedSet, $start, $end));
        }
        return $res;
    }

    /**
     * redis sorted_set zcount
     */
    public function zcount($range_l = '-inf', $range_r = '+inf', $sortedSet = null) {
        $sortedSet = is_null($sortedSet) ? self::$sortedSet : $sortedSet;
        $res = $this->doCmd(self::ZCOUNT, array($sortedSet, $range_l, $range_r));
        return $res;
    }

    //=======================
    //set 接口
    //=======================
    public static $setPool = 'mgset';

    public function changeSet($setPool = 'mqset') {
        self::$setPool = $setPool;
        return $this;
    }

    /**
     * redis set sadd
     */
    public function sadd($member, $expiration = null, $set = null, $asyn = false) {
        try {
            $lifetime = $this->getLifetime($expiration);
            $setPool = is_null($set) ? self::$setPool : $set;
            $res = $this->doCmd(self::SADD, array($setPool, $member), $asyn);
            return $res;
        } catch (Exception $e) {
            crond_log("sadd[{$setPool}]: " . $e->getMessage(), self::LOG_FILE_NAME);
            return false;
        }
    }

    /**
     * redis set spop
     */
    public function spop() {
        try {
            $setPool = self::$setPool;
            $res = $this->doCmd(self::SPOP, array($setPool));
            return $res;
        } catch (Exception $e) {
            crond_log("spop[{$setPool}]: " . $e->getMessage(), self::LOG_FILE_NAME);
            return false;
        }
    }

    /**
     * redis set srem
     */
    public function srem($member, $asyn = false) {
        try {
            $setPool = self::$setPool;
            $res = $this->doCmd(self::SREM, array($setPool, $member), $asyn);
            return $res;
        } catch (Exception $e) {
            crond_log("spop[{$setPool}]: " . $e->getMessage(), self::LOG_FILE_NAME);
            return false;
        }
    }

    /**
     * redis set smembers
     */
    public function smembers() {
        try {
            $setPool = self::$setPool;
            $items = $this->doCmd(self::SMEMBERS, array($setPool));
            if (!is_array($items)) return null;
            return $items;
        } catch (Exception $e) {
            crond_log("smembers[{$setPool}]: " . $e->getMessage(), self::LOG_FILE_NAME);
            return false;
        }
    }

    /**
     * redis set scard
     */
    public function scard() {
        try {
            $setPool = self::$setPool;
            $res = $this->doCmd(self::SCARD, array($setPool));
            return $res;
        } catch (Exception $e) {
            crond_log("[{$setPool}]: " . $e->getMessage(), self::LOG_FILE_NAME);
            return false;
        }
    }
    //======================= Sorted Set End ====================
    //===========================================================

    // 这边对传入的数据进行一个检测
    private function dealParam($param) {
        if (is_numeric($param)) {
            return $param;
        }
        if (mb_strlen($param) >= 10240) {
            // 超过10k数据进行压缩
            $param = gzcompress($param);
            $param = 'T' . $param;
        } else {
            $param = 'F' . $param;
        }
        return $param;
    }

    /**
     * 最后的发送指令和添加扩展协议
     * @param $cmd string 即将发送给go agent的数据
     * @param array $params
     * @param bool $asyn
     * @return array
     */
    public function doCmd($cmd, array $params, $asyn = false) {
        $length = 1 + count($params) + 1; // 第一个是cache实例标志，后一个是扩展的协议数
        $CMD = '*' . $length . self::CRLF;
        $CMD .= $this->preCmd();
        $CMD .= $cmd;
        foreach ($params as $param) {
            $param = $this->dealParam($param);
            $CMD .= '$' . mb_strlen($param) . self::CRLF . $param . self::CRLF;
        }
        if ($asyn) {
            $CMD = 'T' . $CMD;
        } else {
            $CMD = 'F' . $CMD;
        }
        $len = 4 + mb_strlen($CMD);
        $len = sprintf("%04d", $len);
        $CMD = $len . $CMD;

        $res = $this->_conn->call($CMD);
        return $res;
    }

    /**
     * 对指令的前段进行生成
     * @param len int 当前的指令数,默认为1
     * @return string 返回标准的头部
     */
    public function preCmd() {
        $cmd = "";
        foreach ($this->_cache as $value) {
            $cmd .= '$' . mb_strlen($value) . self::CRLF . $value . self::CRLF;
        }
        return $cmd;
    }

    /**
     * 构建多行命令
     * @param $commands array
     * @param bool $asyn
     * @return array
     */
    public function buildMultiCmd(array $commands, $asyn = false) {
        foreach ($commands as $command) {
            if (count($command) === 2) {
                $length = 1 + count($command['params']) + 1; // 第一个是cache实例标志，后一个是扩展的协议数
            } else {
                $length = 2;
            }
            $CMD = '*' . $length . self::CRLF;
            $CMD .= $this->preCmd();
            $CMD .= $command['cmd'];
            if (count($command) === 2) {
                foreach ($command['params'] as $param) {
                    $param = $this->dealParam($param);
                    $CMD .= '$' . mb_strlen($param) . self::CRLF . $param . self::CRLF;
                }
            }
            if ($asyn) {
                $CMD = 'T' . $CMD;
            } else {
                $CMD = 'F' . $CMD;
            }
            $len = 4 + mb_strlen($CMD);
            $len = sprintf("%04d", $len);
            $CMD = $len . $CMD;
            $res[] = $CMD;
        }
        return $res;
    }

    //redis 传入的是过期的时间; Unit: second
    //$lifetime 是过期的时间.
    // Max: 1 week. Min: 60 seconds
    public function getLifetime($lifetime = null) {
        if (is_null($lifetime) || 0 == $lifetime) {
            return 900;
        }

        //正常lifetime
        if ($lifetime < 5184000) {
            //2个月;
            return $lifetime;
        }

        return 900;
    }
}


