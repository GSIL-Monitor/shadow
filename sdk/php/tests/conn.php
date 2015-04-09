<?php

/**
 * Created By VIM
 * @author ziyuan
 * @date January 04, 2015
 * @project shadow
 * @description cache midware
 */
class Conn {
    const UNIX_DOMAIN = "/tmp/shadow_agent.sock";
    /**
     * @var Conn
     */
    private static $_instance = null;
    /**
     * @var Redis  connection to Shadow_Agent
     */
    private $_conn = null;


    /**
     * Keep one instance for performance
     * @return Conn
     */
    public static function build() {
        if (self::$_instance == null) {
            self::$_instance = new self();
        }
        return self::$_instance;
    }

    private function __construct() {
    }

    public function __destruct() {
        $this->close();
    }


    public function connect() {
        $socket = socket_create(AF_UNIX, SOCK_STREAM, 0); //第三个参数为0
        if (FALSE === $socket || $socket < 0) {
            return false;
        }

        $result = socket_connect($socket, Conn::UNIX_DOMAIN); //这里只要两个参数即可
        if (!$result) {
            socket_close($socket);
            return false;
        }

        $this->_conn = $socket;
        return true;
    }

    /**
     * 提交到Agent
     * @param $command string
     * @return array
     */
    public function call($command = null) {
        try {
            $this->connect();
            if (null === $command) {
                return false;
            }

            $s = microtime(true) * 1000 * 10;
            if (!socket_write($this->_conn, $command, strlen($command))) {
                $this->close();
                return false;
            }
            $cost = microtime(true) * 1000 * 10 - $s;
            echo "php call cost , write -- " . $cost . " \n";

            $s = microtime(true) * 1000 * 10;
            $len = socket_read($this->_conn, 5); // 读取前5位长度
            $cost = microtime(true) * 1000 * 10 - $s;
            echo "php call cost , read --- " . $cost . " \n";

            $s = microtime(true) * 1000 * 10;
            $data = Parser::parse($this->_conn, $len);
            $this->close();
            $cost = microtime(true) * 1000 * 10 - $s;
            echo "php call cost , parse +++ " . $cost . " \n";
            return $data;
        } catch (Exception $e) {
            $this->close();
            return false;
        }
    }

    /**
     * 支持Redis的multi模式
     * @param array $commands
     * @internal param array $commands
     * @return array
     */
    public function pipeline(array $commands) {
        try {
            $this->connect();
            $length = count($commands) - 1;
            for ($i = 0; $i < $length; $i++) {
                $command = array_shift($commands);
                socket_write($this->_conn, $command, strlen($command));
                socket_read($this->_conn, 1024, PHP_NORMAL_READ);
                socket_read($this->_conn, 1, PHP_NORMAL_READ); // read the extra \n
            }
            $command = array_shift($commands);
            socket_write($this->_conn, $command, strlen($command));
            $len = socket_read($this->_conn, 5); // 读取前5位长度
            $data = Parser::parse($this->_conn, $len);
            $this->close();
            return $data;
        } catch (Exception $e) {
            $this->close();
            return false;
        }
    }

    public function close() {
        if ($this->_conn != null && is_resource($this->_conn)) {
            socket_close($this->_conn);
            $this->_conn = null;
        }
    }
}
