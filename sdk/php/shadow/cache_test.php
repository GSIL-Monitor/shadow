<?php
require __DIR__ . '/cache.php';

class Shadow_Cache_Test extends PHPUnit_Framework_TestCase {

    /**
     * @before
     */
    public function fixEcho() {
        echo "\n";
    }

    private function buildKey($bytes) {
        return "_FunctionalTesting_" . getmypid() . "_" . $this->build($bytes);
    }

    /**
     * @param $bytes
     * @return string
     */
    private function build($bytes) {
        $characters = '0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz';
        $result = '';
        for ($i = 0; $i < $bytes; $i++) {
            $result .= $characters[mt_rand(0, 61)];
        }
        return $result;
    }

    /**
     * functional testing
     * @throws Exception
     */
    public function testString() {
        $cache = Shadow_Cache::build("mike");

        $simpleKey = "testString_sync";
        $delKey = "testString_sync_delete";

        $s1 = "qwert";
        $s2 = "asdfg";
        $desc = "I am very simple person.  Is it right? Maybe not.";
        $simpleItems = array('HelloWorld' => $s1, 'Meeting' => $s2);
        $complexItems = array(
            'o1' => array("name" => "zhaoxi", "birth-year-month" => 198809),
            'o2' => $desc,
        );

        //=================== sync ============================================
        // delete
        $this->assertTrue($cache->delete("not_existence"));
        $this->assertTrue($cache->delete("not_existence", false));

        // set -> get
        $r = $cache->set($simpleKey, $s1);
        $this->assertTrue($r);
        $r = $cache->get($simpleKey);
        $this->assertEquals($s1, $r);

        // setex -> get
        $r = $cache->set($simpleKey, $s2);
        $this->assertTrue($r);
        $r = $cache->get($simpleKey);
        $this->assertEquals($s2, $r);


        // set -> delete -> get
        $r = $cache->set($delKey, $s1);
        $this->assertTrue($r);
        $r = $cache->delete($delKey);
        $this->assertTrue($r);
        $r = $cache->get($delKey);
        $this->assertEquals(false, $r);


        $r = $cache->set($delKey, $s1);
        $this->assertTrue($r);


        // setBatch/setMulti -> deleteBatch/deleteMulti -> get/getMulti
        $r = $cache->setBatch($complexItems, 300, false);
        foreach ($r as $one) {
            $this->assertTrue($one);
        }
        $r1 = $cache->setMulti($complexItems, 300, false);
        $this->assertEquals($r, $r1);
        $keys = array("o1", "o2");
        $r = $cache->getMulti($keys);
        $this->assertEquals($complexItems, $r);
        $r = $cache->deleteMulti($keys, false);
        foreach ($r as $one) {
            $this->assertTrue($one);
        }
        $r = $cache->getMulti($keys);
        foreach ($r as $one) {
            $this->assertNull($one);
        }
        //=================== async ===========================================
    }

    /**
     * Value Size [1,...,1024]  [1KiB,2KiB,...,1MiB]
     * functional testing
     * @throws Exception
     */
    public function testStringVariousSize() {
        $cache = Shadow_Cache::build("mike");
        $keySize = 10;
        $sizes = array();
        for ($i = 1; $i < 1024; $i++) {
            $sizes[] = $i;
        }
        for ($i = 1; $i <= 1048; $i++) {
            $sizes[] = $i * 1024;
        }

        // hit : miss , 2 : 1
        $key = '';
        $simple = '';
        $complex = array();
        foreach ($sizes as $s) {
            // first set data to Redis
            try {
                $key = $this->buildKey($keySize);
                $simple = $this->build($s);
                $complex['age'] = $s;
                $complex['v'] = $simple;
                $complex['simple'] = true;
                $r = $cache->setex($key, 300, $complex, true);
                $this->assertTrue($r);
            } catch (Exception $e) {
                continue;
            }


            // test get
            for ($i = 0; $i < 2; $i++) {
                try {
                    $getRes = $cache->get($key);
                    $this->assertEquals($complex, $getRes);
                } catch (Exception $e) {
                    echo "==========error==========\n";
                    echo "size: " . $s . "\n";
                    echo "key: " . $key . "\n";
                    $this->assertTrue(false);
                }
            }
            // test get miss
            try {
                $getRes = $cache->get($key . "_missing");
                $this->assertNull($getRes);
            } catch (Exception $e) {
                $this->assertTrue(false);
            }
        }
    }


    /**
     *functional testing
     */
    public function  testKeys() {
        $cache = Shadow_Cache::build("mike");
        $sizes = array();
        for ($i = 1; $i <= 1024; $i++) {
            $sizes[] = $i;
        }
        for ($i = 1; $i < 53; $i++) {
            $sizes[] = 1024 + $i * 20;
        }


        foreach ($sizes as $size) {
            $key = $this->buildKey($size);
            try {
                $r = $cache->setex($key, 300, "key_testing", false);
                $this->assertTrue($r);
                $r = $cache->delete($key);
                $this->assertTrue($r);
            } catch (Exception $e) {
                echo "Have one exception\n";
            }
        }
    }


    public function testList() {
        $instance = Shadow_Cache::build("cars");
    }

    public function testHash() {
        $instance = Shadow_Cache::build("sugar");
    }

    public function testSet() {
        $instance = Shadow_Cache::build("home");
    }
}
