package com.mogujie.shadow.monitor.utils;

import java.util.HashMap;
import java.util.LinkedList;
import java.util.List;
import java.util.regex.Matcher;
import java.util.regex.Pattern;

import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.yaml.snakeyaml.Yaml;

import com.mogujie.shadow.monitor.model.RedisInstanceInfo;
import com.mogujie.tesla.furion.client.FurionClient;
import com.mogujie.tesla.furion.client.FurionClientFactory;
import com.mogujie.tesla.furion.common.model.Configuration;

/**
 * @author ziyuan
 * @date 2/5/15 10:20 AM
 * @project shadow-monitor
 */
public class FurionUtils {
    private static final Logger logger = LoggerFactory.getLogger(FurionUtils.class);

    /**
     * 获得所有的Redis实例信息
     * 
     * @return List<String> 192.168.0.1:6379
     */
    public static List<RedisInstanceInfo> getRedisInstances(String ringName, String appName) {
        List<RedisInstanceInfo> redisInstances = new LinkedList<RedisInstanceInfo>();

        try {
            if (ringName != null && !ringName.isEmpty()) {
                FurionClient furionClient = FurionClientFactory.getFurionClient();
                Configuration configuration = (Configuration) furionClient.getConfiguration(Constans.SHADOW_PREFIX
                        + ringName);
                Yaml yaml = new Yaml();
                List<?> rings = (List<?>) ((HashMap<?, ?>) yaml.load(configuration.getValue())).get("shards");
                for (Object o : rings) {
                    String masterinfo = ((HashMap<?, ?>) o).get("master").toString();
                    List<?> slaveinfo = (List<?>) ((HashMap<?, ?>) o).get("slave");

                    RedisInstanceInfo info = filterData(masterinfo);
                    if (info != null) {
                        redisInstances.add(filterData(masterinfo));
                    }
                    if (slaveinfo != null) {
                        for (Object s : slaveinfo) {
                            info = filterData(s.toString());
                            if (info != null) {
                                redisInstances.add(info);
                            }
                        }
                    }
                }
            }
        } catch (Exception e) {
            logger.error("", e);
        }
        return redisInstances;
    }

    private static RedisInstanceInfo filterData(String conf) {
        if (conf == null)
            return null;
        RedisInstanceInfo instanceInfo = null;
        String regex = "(.*):(\\d+)";
        Pattern pattern = Pattern.compile(regex);
        Matcher matcher = pattern.matcher(conf);
        if (matcher.find()) {
            instanceInfo = new RedisInstanceInfo();
            instanceInfo.setIp(matcher.group(1));
            instanceInfo.setPort(Integer.valueOf(matcher.group(2)));
        }
        return instanceInfo;
    }
}
