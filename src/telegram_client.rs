use reqwest::{multipart, Client, Body, StatusCode};
use tokio::fs::File;
use tokio_util::codec::{BytesCodec, FramedRead};
use serde::Serialize;

#[derive(Serialize)]
pub struct DeleteMessage<'a> {
    pub chat_id: &'a i64,
    pub message_id: &'a i64,
}

#[derive(Serialize)]
pub struct SendMessage<'a> {
    pub chat_id: &'a i64,
    pub text: &'a str,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub reply_to_message_id: Option<i64>,
}

#[derive(Serialize)]
pub struct SendChatAction<'a> {
    pub chat_id: &'a i64,
    pub action: &'a str,
}

#[derive(Serialize)]
pub struct SendVideo<'a> {
    pub chat_id: &'a i64,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub reply_to_message_id: Option<i64>,
    pub video_location: &'a str
}

pub async fn delete_message(token: &str, message: &DeleteMessage<'_>) {
    send_request(token, "deleteMessage", message).await;
}

pub async fn send_message(token: &str, message: &SendMessage<'_>) {
    send_request(token, "sendMessage", message).await;
}

pub async fn send_chat_action(token: &str, message: &SendChatAction<'_>) {
    send_request(token, "sendChatAction", message).await;
}

pub async fn send_request<T>(token: &str, method: &str, payload: &T)
where
    T: Serialize,
{
    let api_endpoint = format!("https://api.telegram.org/bot{}/{}", token, method);
    let client = Client::new();
    let response = client.post(api_endpoint).json(payload).send().await;
    if let Ok(response) = response {
        if response.status() != StatusCode::OK {
            println!("Request failed with status code: {:?}", response.status());
        }
    } else if let Err(err) = response {
        println!("Request error: {:?}", err);
    }
}

pub async fn send_video(token: &str, message: &SendVideo<'_>) {
    let client = reqwest::Client::new();
    let api_endpoint = format!("https://api.telegram.org/bot{}/sendVideo?chat_id={}&reply_to_message_id={}&allow_sending_without_reply=true", token, message.chat_id, message.reply_to_message_id.unwrap_or(-1));

    if let Ok(file) = File::open(message.video_location).await {
        let stream = FramedRead::new(file, BytesCodec::new());
        let file_body = Body::wrap_stream(stream);

        if let Ok(some_file) = multipart::Part::stream(file_body)
            .file_name("video")
            .mime_str("video/mp4")
        {
            let form = multipart::Form::new().part("video", some_file);

            if let Ok(response) = client.post(api_endpoint).multipart(form).send().await {
                let _ = response.text().await;
            }
        }
    }
}

