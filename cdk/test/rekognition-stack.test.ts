import { App } from 'aws-cdk-lib';
import { Template } from 'aws-cdk-lib/assertions';
import { RekognitionStack } from '../lib/rekognition-stack';

describe('RekognitionStack', () => {
  const app = new App();
  const stack = new RekognitionStack(app, 'TestRekognitionStack');
  const template = Template.fromStack(stack);

  test('S3バケットが作成されている', () => {
    template.hasResourceProperties('AWS::S3::Bucket', {
      BucketName: {
        Ref: 'RekognitionBucketName',
      },
    });
  });

  test('VPCが作成されている', () => {
    template.hasResourceProperties('AWS::EC2::VPC', {
      CidrBlock: '10.0.0.0/16',
      EnableDnsHostnames: true,
      EnableDnsSupport: true,
    });
  });

  test('ECSクラスターが作成されている', () => {
    template.hasResourceProperties('AWS::ECS::Cluster', {
      ClusterName: 'RekognitionCluster',
    });
  });

  test('ECSタスク定義が作成されている', () => {
    template.hasResourceProperties('AWS::ECS::TaskDefinition', {
      Family: 'RekognitionTask',
      NetworkMode: 'awsvpc',
      RequiresCompatibilities: ['FARGATE'],
      Cpu: '256',
      Memory: '512',
    });
  });

  test('IAMロールが作成されている', () => {
    template.hasResourceProperties('AWS::IAM::Role', {
      AssumeRolePolicyDocument: {
        Statement: [
          {
            Action: 'sts:AssumeRole',
            Effect: 'Allow',
            Principal: {
              Service: 'ecs-tasks.amazonaws.com',
            },
          },
        ],
      },
    });
  });

  test('Step Functionsステートマシンが作成されている', () => {
    template.hasResourceProperties('AWS::StepFunctions::StateMachine', {
      StateMachineName: 'RekognitionStateMachine',
      DefinitionString: {
        'Fn::Join': [
          '',
          [
            '{"StartAt":"Start","States":{"Start":{"Type":"Pass","Next":"RunRekognition"},"RunRekognition":{"Type":"Task","Resource":"arn:aws:states:::ecs:runTask.sync","Parameters":{"Cluster":"',
            {
              Ref: 'ClusterEB0386A7',
            },
            '","TaskDefinition":"',
            {
              Ref: 'TaskDefTaskDefinition1EDB4A67',
            },
            '","LaunchType":"FARGATE","NetworkConfiguration":{"AwsvpcConfiguration":{"Subnets":["',
            {
              Ref: 'VpcPrivateSubnet1Subnet536B997A',
            },
            '","',
            {
              Ref: 'VpcPrivateSubnet2Subnet3788AAA1',
            },
            '"],"SecurityGroups":["',
            {
              'Fn::GetAtt': ['TaskDefSecurityGroup1C5F0CFF', 'GroupId'],
            },
            '"]}},"Overrides":{"ContainerOverrides":[{"Name":"RekognitionContainer","Environment":[{"Name":"S3_BUCKET","Value":"',
            {
              Ref: 'RekognitionBucketName',
            },
            '"}]}]},"End":true}}}',
          ],
        ],
      },
    });
  });
}); 