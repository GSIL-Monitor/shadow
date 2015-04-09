package com.mogujie.shadow.monitor.utils;

import java.io.File;

/**
 * @author ziyuan
 * @date 2/4/15 5:01 PM
 * @project shadow-monitor
 */
public class Constans {

    public static final String CONF_PATH = System.getProperty("user.home") + File.separator + ".shadow-conf";
    public static final String SHADOW_PREFIX = "Shadow.Ring.";
    public static final String SHADOW_SUFFIX = "__Shadow";

    public static void main(String[] args) {
        System.out.println(CONF_PATH);
    }
}
