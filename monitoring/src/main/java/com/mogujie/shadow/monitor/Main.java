package com.mogujie.shadow.monitor;

import java.util.List;
import java.util.Timer;
import java.util.TimerTask;

import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import com.alibaba.fastjson.JSON;
import com.alibaba.fastjson.JSONObject;
import com.mogujie.shadow.monitor.collector.RedisInfoCollector;
import com.mogujie.shadow.monitor.collector.ShadowDataHolder;
import com.mogujie.shadow.monitor.furion.ShadowFurionMock;
import com.mogujie.shadow.monitor.model.RedisInstanceInfo;
import com.mogujie.shadow.monitor.model.ShadowRedisInfo;
import com.mogujie.shadow.monitor.utils.ConfUtils;
import com.mogujie.shadow.monitor.utils.FurionUtils;

/**
 * @author ziyuan
 * @date 2/4/15 4:10 PM
 * @project shadow-monitor
 */
public class Main {

    private static final Logger logger = LoggerFactory.getLogger(Main.class);

    private static TimerTask fetchRedisInfoTask = new TimerTask() {
        @Override
        public void run() {
            try {
                List<String> configsKeys = ConfUtils.readFromConf();

                for (String key : configsKeys) {
                    if (ShadowDataHolder.getInstance().getFurionData(key) != null) {
                        JSONObject jsonObject = JSON.parseObject(ShadowDataHolder.getInstance().getFurionData(key));
                        List<RedisInstanceInfo> instanceInfos = FurionUtils.getRedisInstances(jsonObject.get("value")
                                .toString(), key);
                        for (RedisInstanceInfo redisInstanceInfo : instanceInfos) {
                            RedisInfoCollector collector = new RedisInfoCollector(key);
                            ShadowRedisInfo redisInfo = collector.collect(redisInstanceInfo);
                            logger.info(redisInfo.toString());
                            // 这边加上sentry逻辑
                        }
                    }
                }
            } catch (Exception e) {
                logger.error("", e);
            }
        }
    };

    static Timer fetchRedisInfoTimer = new Timer("fetch redis info timer");

    static final long AGAIN_TIME = 5000; // 重复时间
    static final long DELAY_TIME = 5000; // 因为conf异步拉取，所以这边需要延迟开始

    public static void main(String[] args) {
        ShadowFurionMock shadowFurionMock = new ShadowFurionMock();
        shadowFurionMock.startListen();

        fetchRedisInfoTimer.schedule(fetchRedisInfoTask, DELAY_TIME, AGAIN_TIME);
        logger.info("Bootstrap OK!");
    }
}
