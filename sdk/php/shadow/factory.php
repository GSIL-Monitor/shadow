<?php

/**
 * Created by PhpStorm.
 * User: zhaoxi (zhaoxi@mogujie.com)
 * Date: 15-02-27
 * Time: 16:25
 */
interface Shadow_Factory {
    /**
     * @param $app , AppName that is string
     * @return Shadow_Cache
     * @throws Exception
     */
    static function build($app);
}