import * as cdk from 'aws-cdk-lib';
import * as s3 from 'aws-cdk-lib/aws-s3';
import * as stepfunctions from 'aws-cdk-lib/aws-stepfunctions';
import * as tasks from 'aws-cdk-lib/aws-stepfunctions-tasks';
import * as ecs from 'aws-cdk-lib/aws-ecs';
import * as ec2 from 'aws-cdk-lib/aws-ec2';
import * as iam from 'aws-cdk-lib/aws-iam';
import { Construct } from 'constructs';

export class RekognitionStack extends cdk.Stack {
  constructor(scope: Construct, id: string, props?: cdk.StackProps) {
    super(scope, id, props);

    // S3バケットの作成（画像の保存用）
    const bucket = new s3.Bucket(this, 'RekognitionBucket', {
      removalPolicy: cdk.RemovalPolicy.DESTROY, // スタック削除時にバケットも削除
      autoDeleteObjects: true, // バケット内のオブジェクトも自動削除
    });

    // VPCの作成（ECSクラスター用）
    const vpc = new ec2.Vpc(this, 'Vpc', {
      maxAzs: 2, // 2つのアベイラビリティゾーンを使用
    });

    // ECSクラスターの作成
    const cluster = new ecs.Cluster(this, 'Cluster', {
      vpc,
    });

    // ECSタスク定義（Fargate使用）
    const taskDefinition = new ecs.FargateTaskDefinition(this, 'TaskDef', {
      memoryLimitMiB: 512, // メモリ制限
      cpu: 256, // CPU制限
    });

    // コンテナの定義
    taskDefinition.addContainer('RekognitionContainer', {
      image: ecs.ContainerImage.fromRegistry('your-ecr-repository'), // ECRリポジトリのイメージを使用
      environment: {
        S3_BUCKET: bucket.bucketName, // S3バケット名を環境変数として設定
      },
    });

    // IAMロールの作成（ECSタスク用）
    const taskRole = new iam.Role(this, 'TaskRole', {
      assumedBy: new iam.ServicePrincipal('ecs-tasks.amazonaws.com'),
    });

    // 必要な権限の付与（最小権限の原則に基づく）
    taskRole.addToPolicy(new iam.PolicyStatement({
      actions: [
        'rekognition:CompareFaces', // Rekognitionの顔比較
        's3:GetObject', // S3からのオブジェクト取得
        's3:PutObject', // S3へのオブジェクト保存
      ],
      resources: ['*'],
    }));

    // Step Functionsの定義
    const definition = new stepfunctions.Pass(this, 'Start')
      .next(new tasks.EcsRunTask(this, 'RunRekognition', {
        integrationPattern: stepfunctions.IntegrationPattern.RUN_JOB, // ジョブ実行パターン
        cluster,
        taskDefinition,
        launchTarget: new tasks.EcsFargateLaunchTarget(), // Fargate起動
        containerOverrides: [{
          containerName: 'RekognitionContainer',
          environment: [
            {
              name: 'S3_BUCKET',
              value: bucket.bucketName,
            },
          ],
        }],
      }));

    // Step Functionsステートマシンの作成
    new stepfunctions.StateMachine(this, 'RekognitionStateMachine', {
      definition,
      timeout: cdk.Duration.minutes(5), // タイムアウト設定
    });
  }
} 