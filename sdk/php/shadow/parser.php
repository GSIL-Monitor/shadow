<?php
   /**
   * Created by VIM & PHPSTORM
   * @author: zhaoxi(LinZuxiong), ziyuan
   * Date: 14-12-29
   * Time: 14:55
   */
   class Parser {

      const STATUS_REPLY     = '+';
      const ERROR_REPLY      = '-';
      const INTEGER_REPLY    = ':';
      const BULK_REPLY       = '$';
      const MULTI_BULK_REPLY = '*';

      /**
      * 解析redis协议  TODO zhaoxi must *throw exception *logging exception
      */
      public static function parse($socket = null, $len = 0) {
         try{
            if (null === $socket) {
               return false;
            }
            // echo "Read From Server...\n";
            $reply = socket_read($socket, 1024, PHP_NORMAL_READ);
            socket_read($socket, 1, PHP_NORMAL_READ);

            if (FALSE === $reply || strlen($reply) == 0) return false;

            $reply = trim($reply);
            $reply_type = $reply[0];
            $data = substr($reply, 1);
            switch($reply_type) {
               case Parser::STATUS_REPLY:
                  $res = true;
                  break;

               case Parser::ERROR_REPLY:
                  $res = false;
                  break;

               case Parser::INTEGER_REPLY:
                  $res = intval($data);
                  break;

               case Parser::BULK_REPLY:
                  $datalength = intval($data);
                  if ($datalength < 0) return NULL;
                  $bulkreply = socket_read($socket, $datalength + 1, PHP_NORMAL_READ);
                  socket_read($socket, 1);
                  if (FALSE === $bulkreply) $res = false;
                  $bulkreply = trim($bulkreply);

                  if (substr($bulkreply, 0, 1) == 'T') {
                     $bulkreply = gzuncompress(substr($bulkreply, 1, strlen($bulkreply)));
                  }else {
                     $bulkreply = substr($bulkreply, 1, strlen($bulkreply));
                  }
                  $res = $bulkreply;
                  break;

               case Parser::MULTI_BULK_REPLY:
                  $bulkreplycount = intval($data);
                  if ($bulkreplycount < 0) return NULL;
                  $multibulkreply = array();
                  for($i = 0; $i < $bulkreplycount; $i++){
                     $multibulkreply[] = Parser::parse($socket, $len = 0);
                  }
                  $res = $multibulkreply;
                  break;

               default:
                  throw new Exception("Unknown Reply Type: $reply");
                  break;
               }
               return $res;
            }catch(Exception $e){
               return false;
            }
         }
      }
