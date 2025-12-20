import json
import boto3
import os
from boto3.dynamodb.conditions import Key, Attr

dynamodb = boto3.resource('dynamodb')
table = dynamodb.Table('web_socket_chat_connections')
chat_table = dynamodb.Table('fre_style_chat')


def lambda_handler(event, context):
    route_key = event['requestContext']['routeKey']
    connection_id = event['requestContext']['connectionId']
    domain_name = event['requestContext']['domainName']
    stage = event['requestContext']['stage']

    # lambdaからAPI Gateway management APIへメッセージを送るためのAWS API
    apigw_management = boto3.client(
        'apigatewaymanagementapi',
        endpoint_url=f"https://{domain_name}/{stage}"
    )

    # --- 接続処理 ---
    if route_key == '$connect':
        params = event.get('queryStringParameters') or {}
        user_id = params.get('user_id', 'anonymous')
        room_id = params.get('room_id', 'default')

        table.put_item(Item={
            'connection_id': connection_id,
            'user_id': user_id,
            'room_id': room_id
        })

        return {'statusCode': 200}

    # --- 切断処理 ---
    elif route_key == '$disconnect':
        table.delete_item(Key={
            'connection_id': connection_id
        })

        return {'statusCode': 200}

    # --- メッセージ送信処理 ($default) ---
    elif route_key == '$default':
        print("💬 $default route triggered")

        # --- 受信データを取得 ---
        body = json.loads(event.get('body', '{}'))
        action = body.get('action')
        
        # --- メッセージ削除処理 ---
        if action == 'delete':
            print("🗑️ Delete action triggered")
            room_id = body.get('room_id')
            timestamp = body.get('timestamp')
            sender_id = body.get('sender_id')
            
            if not all([room_id, timestamp, sender_id]):
                return {
                    'statusCode': 400,
                    'body': json.dumps({'error': 'room_id, timestamp, sender_id required'})
                }
            
            try:
                # --- DynamoDBからメッセージを削除 ---
                print(f'🗑️ Deleting message: room_id={room_id}, timestamp={timestamp}, type(timestamp)={type(timestamp).__name__}')
                
                # timestampが文字列の場合は数値に変換
                try:
                    timestamp_num = float(timestamp) if isinstance(timestamp, str) else timestamp
                except (ValueError, TypeError):
                    print(f'❌ タイムスタンプ変換失敗: {timestamp}')
                    return {
                        'statusCode': 400,
                        'body': json.dumps({'error': 'Invalid timestamp format'})
                    }
                
                response = chat_table.delete_item(
                    Key={
                        'room_id': int(room_id),
                        'timestamp': timestamp_num
                    }
                )
                
                print(f'✅ DynamoDB削除応答: {response}')
                
                # --- 同じルームにいる全員に削除通知を送信 ---
                response = table.scan(
                    FilterExpression=Attr('room_id').eq(room_id)
                )
                
                for item in response['Items']:
                    target_conn_id = item['connection_id']
                    
                    try:
                        apigw_management.post_to_connection(
                            ConnectionId=target_conn_id,
                            Data=json.dumps({
                                'type': 'message_deleted',
                                'room_id': room_id,
                                'timestamp': timestamp_num,
                                'sender_id': sender_id
                            }).encode('utf-8')
                        )
                    except apigw_management.exceptions.GoneException:
                        table.delete_item(Key={'connection_id': target_conn_id})
                
                print(f'✅ メッセージ削除完了: room_id={room_id}, timestamp={timestamp_num}')
                return {'statusCode': 200, 'body': json.dumps({'success': True})}
            
            except Exception as e:
                print(f'❌ メッセージ削除エラー: {str(e)}')
                import traceback
                traceback.print_exc()
                return {
                    'statusCode': 500,
                    'body': json.dumps({'error': str(e)})
                }
        
        # --- 通常のメッセージ送信処理 ---
        room_id = body.get('room_id')
        sender_id = body.get('sender_id')
        content = body.get('content')

        if not all([room_id, sender_id, content]):
            return {
                'statusCode': 400,
                'body': json.dumps({'error': 'Invalid message format'})
            }

        # --- DynamoDBに接続情報を確認 ---
        conn_data = table.get_item(Key={'connection_id': connection_id})
        if 'Item' not in conn_data:
            return {'statusCode': 400, 'body': 'Connection not found'}

        # --- 同じ room_id の接続一覧を取得 ---
        response = table.scan(
            FilterExpression=Attr('room_id').eq(room_id)
        )

        # --- DynamoDB（チャット履歴テーブル）に保存 ---
        timestamp = int(context.aws_request_id[-6:], 16)
        chat_table.put_item(
            Item={
                'room_id': int(room_id),
                'timestamp': timestamp,
                'sender_id': sender_id,
                'content': content
            }
        )

        # --- 同じルームにいる全員にメッセージ送信 ---
        for item in response['Items']:
            target_conn_id = item['connection_id']

            if target_conn_id == connection_id:
                continue

            try:
                apigw_management.post_to_connection(
                    ConnectionId=target_conn_id,
                    Data=json.dumps({
                        'room_id': room_id,
                        'sender_id': sender_id,
                        'content': content,
                        'timestamp': timestamp
                    }).encode('utf-8')
                )
            except apigw_management.exceptions.GoneException:
                # 切断済みの接続を削除
                table.delete_item(Key={'connection_id': target_conn_id})

        return {'statusCode': 200}

    # --- 無効なルート ---
    else:
        return {'statusCode': 400, 'body': 'Invalid route'}
