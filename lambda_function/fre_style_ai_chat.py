import json
import time
import boto3
from botocore.exceptions import ClientError
from boto3.dynamodb.conditions import Key, Attr


# Bedrock Client
bedrock = boto3.client("bedrock-runtime", region_name="ap-northeast-1")

# DynamoDB（接続管理、履歴保存）
dynamodb = boto3.resource("dynamodb")
conn_table = dynamodb.Table('web_socket_chat_connection') # WebSocket用
history_table = dynamodb.Table("fre_style_ai_chat") # チャット履歴用



# モデルID
MODEL_ID = "anthropic.claude-3-haiku-20240307-v1:0"

def lambda_handler(event, context):
  route_key = event["requestContext"]["routeKey"]
  connection_id = event["requestContext"]["connectionId"]
  domain_name = event["requestContext"]["domainName"]
  stage = event["requestContext"]["stage"]
  
# WebSocket用の送信用クライアント
# ここでAPI Gateway(WebSocketサーバー)側にメッセージを送る
  apigw_management = boto3.client(
     "apigatewaymanagementapi",
      endpoint_url=f"https://{domain_name}/{stage}"
)

  if route_key == "$default":
    body = json.loads(event.get("body", "{}"))
    sender_id = body.get("sender_id")
    user_message = body.get("content")
  
    if not sender_id or not user_message:
      return {
        "statusCode": 400,
        "body": "sender_id and content required"
      }
      
    current_time = int(time.time())
    prompt = f"Human:{user_message}\n\nAssistant:"
    
    try:
        # Bedrock呼び出し
        response = bedrock.invoke_model(
            modelId=MODEL_ID,
            contentType="application/json",
            accept="application/json",
            body=json.dumps({
                "anthropic_version": "bedrock-2023-05-31",
                "max_tokens": 500,
                "temperature":0.7,
                "messages": [
                  {
                    "role": "user",
                    "content": [
                      {
                        "type": "text",
                        "text": user_message
                      }
                    ]
                  }
                ]
            })
        )
        response_body = json.loads(response["body"].read())
        ai_reply = response_body["content"][0]["text"].strip()
        
        
        # 履歴を保存する（人間のプロンプト）
        history_table.put_item(Item={
            "sender_id": sender_id,
            "timestamp": current_time,
            "content": user_message,
            "is_user": True
        })
        
        # AIへの応答
        history_table.put_item(Item={
            "sender_id": sender_id,
            "timestamp": current_time + 1,
            "content": ai_reply,
            "is_user": False
        })
        
        # 応答をWebSocketでクライアントに送信をする
        # ユーザーとAIでタイムスタンプが一緒だとソートが崩れる
        apigw_management.post_to_connection(
            ConnectionId=connection_id,
            Data=json.dumps({
                "reply": ai_reply,
                "from": "AI",
                "timestamp": current_time + 1
            }).encode('utf-8')
        )
        
        return { 
          "statusCode": 200 ,
          "body": ai_reply}
    except apigw_management.exceptions.GoneException:
        # クライアントが切断していたら接続情報削除する
        conn_table.delete_item(Key={'connection_id':connection_id})
        return { "statusCode": 410 }
    
    except Exception as e:
        print("Error:", str(e))
        return {
            "statusCode": 500,
            "body": json.dumps({"error": str(e)})
        }
  else:
    return { "statusCode": 400, "body": "Invalid route" }