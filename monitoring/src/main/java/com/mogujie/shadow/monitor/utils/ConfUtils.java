package com.mogujie.shadow.monitor.utils;

import java.io.BufferedReader;
import java.io.File;
import java.io.FileReader;
import java.io.IOException;
import java.util.LinkedList;
import java.util.List;

import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

/**
 * @author ziyuan
 * @date 2/5/15 9:52 AM
 * @project shadow-monitor
 */
public class ConfUtils {
    private static final Logger logger = LoggerFactory.getLogger(ConfUtils.class);

    /**
     * 从配置文件中读取配置的APPNAMES
     * 
     * @return List
     * @throws java.io.IOException
     */
    public static List<String> readFromConf() {
        List<String> appnames = new LinkedList<String>();
        BufferedReader reader = null;
        try {
            File confFile = new File(Constans.CONF_PATH);
            if (!confFile.exists()) {
                confFile.createNewFile();
            }

            reader = new BufferedReader(new FileReader(confFile));
            String data = "";
            while ((data = reader.readLine()) != null) {
                if (!data.equals("")) {
                    appnames.add(data);
                }
            }
        } catch (Exception e) {
            logger.error("", e);
        } finally {
            if (reader != null) {
                try {
                    reader.close();
                } catch (IOException e) {
                }
            }
        }
        return appnames;
    }
}
