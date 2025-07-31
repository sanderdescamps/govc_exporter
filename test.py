send_targets = """[
            {
              "address": "10.136.15.233",
              "port": 3260,
              "authenticationProperties": {
                "chapAuthEnabled": false,
                "chapAuthenticationType": "chapProhibited",
                "chapInherited": true,
                "mutualChapAuthenticationType": "chapProhibited",
                "mutualChapInherited": true
              },
              "digestProperties": {
                "headerDigestType": "digestProhibited",
                "headerDigestInherited": true,
                "dataDigestType": "digestProhibited",
                "dataDigestInherited": true
              },
              "supportedAdvancedOptions": [
                {
                  "label": "ErrorRecoveryLevel",
                  "summary": "iSCSI option : iSCSI Error Recovery Level (ERL) value that the ESX initiator would negotiate during login.",
                  "key": "ErrorRecoveryLevel",
                  "optionType": {
                    "valueIsReadonly": true,
                    "min": 0,
                    "max": 2,
                    "defaultValue": 0
                  }
                },
                {
                  "label": "LoginRetryMax",
                  "summary": "iSCSI option : Maximum number of times ESX initiator would retry login to a target, before giving up.",
                  "key": "LoginRetryMax",
                  "optionType": {
                    "valueIsReadonly": true,
                    "min": 0,
                    "max": 64,
                    "defaultValue": 4
                  }
                },
                {
                  "label": "MaxOutstandingR2T",
                  "summary": "iSCSI option : Maximum number of R2T (Ready To Transfer) PDUs, that can be outstanding for a task.",
                  "key": "MaxOutstandingR2T",
                  "optionType": {
                    "valueIsReadonly": false,
                    "min": 1,
                    "max": 8,
                    "defaultValue": 1
                  }
                },
                {
                  "label": "FirstBurstLength",
                  "summary": "iSCSI option : Maximum unsolicited data in bytes initiator can send during the execution of a single SCSI command. It must be <= MaxBurstLength.",
                  "key": "FirstBurstLength",
                  "optionType": {
                    "valueIsReadonly": false,
                    "min": 512,
                    "max": 16777215,
                    "defaultValue": 262144
                  }
                },
                {
                  "label": "MaxBurstLength",
                  "summary": "iSCSI option : Maximum SCSI data payload in bytes in a Data-In or a solicited Data-Out iSCSI sequence.",
                  "key": "MaxBurstLength",
                  "optionType": {
                    "valueIsReadonly": false,
                    "min": 512,
                    "max": 16777215,
                    "defaultValue": 262144
                  }
                },
                {
                  "label": "MaxRecvDataSegLen",
                  "summary": "iSCSI option : Maximum data segment length in bytes that can be received in an iSCSI PDU. It is recommended to keep it <= MaxBurstLength.",
                  "key": "MaxRecvDataSegLen",
                  "optionType": {
                    "valueIsReadonly": false,
                    "min": 512,
                    "max": 16777215,
                    "defaultValue": 131072
                  }
                },
                {
                  "label": "MaxCommands",
                  "summary": "iSCSI option : Maximum SCSI commands that can be queued on the iscsi adpater.",
                  "key": "MaxCommands",
                  "optionType": {
                    "valueIsReadonly": true,
                    "min": 2,
                    "max": 2048,
                    "defaultValue": 128
                  }
                },
                {
                  "label": "DefaultTimeToWait",
                  "summary": "iSCSI option : Minimum seconds to wait before attempting a logout or an active task reassignment after an unexpected connection termination or reset.",
                  "key": "DefaultTimeToWait",
                  "optionType": {
                    "valueIsReadonly": true,
                    "min": 0,
                    "max": 60,
                    "defaultValue": 2
                  }
                },
                {
                  "label": "DefaultTimeToRetain",
                  "summary": "iSCSI option : Maximum seconds that a connection and task allegiance reinstatement is still possible following a connection termination or reset.",
                  "key": "DefaultTimeToRetain",
                  "optionType": {
                    "valueIsReadonly": true,
                    "min": 0,
                    "max": 60,
                    "defaultValue": 0
                  }
                },
                {
                  "label": "LoginTimeout",
                  "summary": "iSCSI option : Time in seconds initiator will wait for the Login response to finish.",
                  "key": "LoginTimeout",
                  "optionType": {
                    "valueIsReadonly": false,
                    "min": 1,
                    "max": 60,
                    "defaultValue": 5
                  }
                },
                {
                  "label": "LogoutTimeout",
                  "summary": "iSCSI option : Time in seconds initiator will wait to get a response for Logout request PDU",
                  "key": "LogoutTimeout",
                  "optionType": {
                    "valueIsReadonly": true,
                    "min": 0,
                    "max": 60,
                    "defaultValue": 15
                  }
                },
                {
                  "label": "RecoveryTimeout",
                  "summary": "iSCSI option : Time in seconds that can elapse while a session recovery is performed. If the session recovery is not over within this time, initiator terminates the session.",
                  "key": "RecoveryTimeout",
                  "optionType": {
                    "valueIsReadonly": false,
                    "min": 1,
                    "max": 120,
                    "defaultValue": 10
                  }
                },
                {
                  "label": "NoopTimeout",
                  "summary": "iSCSI option : Time in seconds that can elapse before initiator receives a NOP-IN PDU from the target. When no-op timeout limit is exceeded, initiator terminates the current connection and starts a new one.",
                  "key": "NoopTimeout",
                  "optionType": {
                    "valueIsReadonly": false,
                    "min": 10,
                    "max": 30,
                    "defaultValue": 10
                  }
                },
                {
                  "label": "NoopInterval",
                  "summary": "iSCSI option : Time in seconds in between NOP-OUTs sent by the initiator to verify that a connection is still active.",
                  "key": "NoopInterval",
                  "optionType": {
                    "valueIsReadonly": false,
                    "min": 1,
                    "max": 60,
                    "defaultValue": 15
                  }
                },
                {
                  "label": "InitR2T",
                  "summary": "iSCSI option : Whether to allow initiator to start sending data to a target as if it had received an initial R2T.",
                  "key": "InitR2T",
                  "optionType": {
                    "valueIsReadonly": true,
                    "supported": true,
                    "defaultValue": false
                  }
                },
                {
                  "label": "ImmediateData",
                  "summary": "iSCSI option : Whether to allow initiator to send immediate (unsolicited) data to the target.",
                  "key": "ImmediateData",
                  "optionType": {
                    "valueIsReadonly": true,
                    "supported": true,
                    "defaultValue": true
                  }
                },
                {
                  "label": "DelayedAck",
                  "summary": "iSCSI option : Whether to allow initiator to delay acknowledgement of received data packets.",
                  "key": "DelayedAck",
                  "optionType": {
                    "valueIsReadonly": false,
                    "supported": true,
                    "defaultValue": true
                  }
                }
              ],
              "advancedOptions": [
                {
                  "key": "ErrorRecoveryLevel",
                  "value": 0,
                  "isInherited": true
                },
                {
                  "key": "LoginRetryMax",
                  "value": 4,
                  "isInherited": true
                },
                {
                  "key": "MaxOutstandingR2T",
                  "value": 1,
                  "isInherited": true
                },
                {
                  "key": "FirstBurstLength",
                  "value": 262144,
                  "isInherited": true
                },
                {
                  "key": "MaxBurstLength",
                  "value": 262144,
                  "isInherited": true
                },
                {
                  "key": "MaxRecvDataSegLen",
                  "value": 131072,
                  "isInherited": true
                },
                {
                  "key": "MaxCommands",
                  "value": 128,
                  "isInherited": true
                },
                {
                  "key": "DefaultTimeToWait",
                  "value": 0,
                  "isInherited": true
                },
                {
                  "key": "DefaultTimeToRetain",
                  "value": 0,
                  "isInherited": true
                },
                {
                  "key": "LoginTimeout",
                  "value": 5,
                  "isInherited": true
                },
                {
                  "key": "LogoutTimeout",
                  "value": 15,
                  "isInherited": true
                },
                {
                  "key": "RecoveryTimeout",
                  "value": 10,
                  "isInherited": true
                },
                {
                  "key": "NoopTimeout",
                  "value": 10,
                  "isInherited": true
                },
                {
                  "key": "NoopInterval",
                  "value": 15,
                  "isInherited": true
                },
                {
                  "key": "InitR2T",
                  "value": false,
                  "isInherited": true
                },
                {
                  "key": "ImmediateData",
                  "value": false,
                  "isInherited": true
                },
                {
                  "key": "DelayedAck",
                  "value": true,
                  "isInherited": true
                }
              ],
              "parent": "vmhba64"
            },
            {
              "address": "10.136.13.231",
              "port": 3260,
              "authenticationProperties": {
                "chapAuthEnabled": false,
                "chapAuthenticationType": "chapProhibited",
                "chapInherited": true,
                "mutualChapAuthenticationType": "chapProhibited",
                "mutualChapInherited": true
              },
              "digestProperties": {
                "headerDigestType": "digestProhibited",
                "headerDigestInherited": true,
                "dataDigestType": "digestProhibited",
                "dataDigestInherited": true
              },
              "supportedAdvancedOptions": [
                {
                  "label": "ErrorRecoveryLevel",
                  "summary": "iSCSI option : iSCSI Error Recovery Level (ERL) value that the ESX initiator would negotiate during login.",
                  "key": "ErrorRecoveryLevel",
                  "optionType": {
                    "valueIsReadonly": true,
                    "min": 0,
                    "max": 2,
                    "defaultValue": 0
                  }
                },
                {
                  "label": "LoginRetryMax",
                  "summary": "iSCSI option : Maximum number of times ESX initiator would retry login to a target, before giving up.",
                  "key": "LoginRetryMax",
                  "optionType": {
                    "valueIsReadonly": true,
                    "min": 0,
                    "max": 64,
                    "defaultValue": 4
                  }
                },
                {
                  "label": "MaxOutstandingR2T",
                  "summary": "iSCSI option : Maximum number of R2T (Ready To Transfer) PDUs, that can be outstanding for a task.",
                  "key": "MaxOutstandingR2T",
                  "optionType": {
                    "valueIsReadonly": false,
                    "min": 1,
                    "max": 8,
                    "defaultValue": 1
                  }
                },
                {
                  "label": "FirstBurstLength",
                  "summary": "iSCSI option : Maximum unsolicited data in bytes initiator can send during the execution of a single SCSI command. It must be <= MaxBurstLength.",
                  "key": "FirstBurstLength",
                  "optionType": {
                    "valueIsReadonly": false,
                    "min": 512,
                    "max": 16777215,
                    "defaultValue": 262144
                  }
                },
                {
                  "label": "MaxBurstLength",
                  "summary": "iSCSI option : Maximum SCSI data payload in bytes in a Data-In or a solicited Data-Out iSCSI sequence.",
                  "key": "MaxBurstLength",
                  "optionType": {
                    "valueIsReadonly": false,
                    "min": 512,
                    "max": 16777215,
                    "defaultValue": 262144
                  }
                },
                {
                  "label": "MaxRecvDataSegLen",
                  "summary": "iSCSI option : Maximum data segment length in bytes that can be received in an iSCSI PDU. It is recommended to keep it <= MaxBurstLength.",
                  "key": "MaxRecvDataSegLen",
                  "optionType": {
                    "valueIsReadonly": false,
                    "min": 512,
                    "max": 16777215,
                    "defaultValue": 131072
                  }
                },
                {
                  "label": "MaxCommands",
                  "summary": "iSCSI option : Maximum SCSI commands that can be queued on the iscsi adpater.",
                  "key": "MaxCommands",
                  "optionType": {
                    "valueIsReadonly": true,
                    "min": 2,
                    "max": 2048,
                    "defaultValue": 128
                  }
                },
                {
                  "label": "DefaultTimeToWait",
                  "summary": "iSCSI option : Minimum seconds to wait before attempting a logout or an active task reassignment after an unexpected connection termination or reset.",
                  "key": "DefaultTimeToWait",
                  "optionType": {
                    "valueIsReadonly": true,
                    "min": 0,
                    "max": 60,
                    "defaultValue": 2
                  }
                },
                {
                  "label": "DefaultTimeToRetain",
                  "summary": "iSCSI option : Maximum seconds that a connection and task allegiance reinstatement is still possible following a connection termination or reset.",
                  "key": "DefaultTimeToRetain",
                  "optionType": {
                    "valueIsReadonly": true,
                    "min": 0,
                    "max": 60,
                    "defaultValue": 0
                  }
                },
                {
                  "label": "LoginTimeout",
                  "summary": "iSCSI option : Time in seconds initiator will wait for the Login response to finish.",
                  "key": "LoginTimeout",
                  "optionType": {
                    "valueIsReadonly": false,
                    "min": 1,
                    "max": 60,
                    "defaultValue": 5
                  }
                },
                {
                  "label": "LogoutTimeout",
                  "summary": "iSCSI option : Time in seconds initiator will wait to get a response for Logout request PDU",
                  "key": "LogoutTimeout",
                  "optionType": {
                    "valueIsReadonly": true,
                    "min": 0,
                    "max": 60,
                    "defaultValue": 15
                  }
                },
                {
                  "label": "RecoveryTimeout",
                  "summary": "iSCSI option : Time in seconds that can elapse while a session recovery is performed. If the session recovery is not over within this time, initiator terminates the session.",
                  "key": "RecoveryTimeout",
                  "optionType": {
                    "valueIsReadonly": false,
                    "min": 1,
                    "max": 120,
                    "defaultValue": 10
                  }
                },
                {
                  "label": "NoopTimeout",
                  "summary": "iSCSI option : Time in seconds that can elapse before initiator receives a NOP-IN PDU from the target. When no-op timeout limit is exceeded, initiator terminates the current connection and starts a new one.",
                  "key": "NoopTimeout",
                  "optionType": {
                    "valueIsReadonly": false,
                    "min": 10,
                    "max": 30,
                    "defaultValue": 10
                  }
                },
                {
                  "label": "NoopInterval",
                  "summary": "iSCSI option : Time in seconds in between NOP-OUTs sent by the initiator to verify that a connection is still active.",
                  "key": "NoopInterval",
                  "optionType": {
                    "valueIsReadonly": false,
                    "min": 1,
                    "max": 60,
                    "defaultValue": 15
                  }
                },
                {
                  "label": "InitR2T",
                  "summary": "iSCSI option : Whether to allow initiator to start sending data to a target as if it had received an initial R2T.",
                  "key": "InitR2T",
                  "optionType": {
                    "valueIsReadonly": true,
                    "supported": true,
                    "defaultValue": false
                  }
                },
                {
                  "label": "ImmediateData",
                  "summary": "iSCSI option : Whether to allow initiator to send immediate (unsolicited) data to the target.",
                  "key": "ImmediateData",
                  "optionType": {
                    "valueIsReadonly": true,
                    "supported": true,
                    "defaultValue": true
                  }
                },
                {
                  "label": "DelayedAck",
                  "summary": "iSCSI option : Whether to allow initiator to delay acknowledgement of received data packets.",
                  "key": "DelayedAck",
                  "optionType": {
                    "valueIsReadonly": false,
                    "supported": true,
                    "defaultValue": true
                  }
                }
              ],
              "advancedOptions": [
                {
                  "key": "ErrorRecoveryLevel",
                  "value": 0,
                  "isInherited": true
                },
                {
                  "key": "LoginRetryMax",
                  "value": 4,
                  "isInherited": true
                },
                {
                  "key": "MaxOutstandingR2T",
                  "value": 1,
                  "isInherited": true
                },
                {
                  "key": "FirstBurstLength",
                  "value": 262144,
                  "isInherited": true
                },
                {
                  "key": "MaxBurstLength",
                  "value": 262144,
                  "isInherited": true
                },
                {
                  "key": "MaxRecvDataSegLen",
                  "value": 131072,
                  "isInherited": true
                },
                {
                  "key": "MaxCommands",
                  "value": 128,
                  "isInherited": true
                },
                {
                  "key": "DefaultTimeToWait",
                  "value": 0,
                  "isInherited": true
                },
                {
                  "key": "DefaultTimeToRetain",
                  "value": 0,
                  "isInherited": true
                },
                {
                  "key": "LoginTimeout",
                  "value": 5,
                  "isInherited": true
                },
                {
                  "key": "LogoutTimeout",
                  "value": 15,
                  "isInherited": true
                },
                {
                  "key": "RecoveryTimeout",
                  "value": 10,
                  "isInherited": true
                },
                {
                  "key": "NoopTimeout",
                  "value": 10,
                  "isInherited": true
                },
                {
                  "key": "NoopInterval",
                  "value": 15,
                  "isInherited": true
                },
                {
                  "key": "InitR2T",
                  "value": false,
                  "isInherited": true
                },
                {
                  "key": "ImmediateData",
                  "value": false,
                  "isInherited": true
                },
                {
                  "key": "DelayedAck",
                  "value": true,
                  "isInherited": true
                }
              ],
              "parent": "vmhba64"
            },
            {
              "address": "10.136.13.230",
              "port": 3260,
              "authenticationProperties": {
                "chapAuthEnabled": false,
                "chapAuthenticationType": "chapProhibited",
                "chapInherited": true,
                "mutualChapAuthenticationType": "chapProhibited",
                "mutualChapInherited": true
              },
              "digestProperties": {
                "headerDigestType": "digestProhibited",
                "headerDigestInherited": true,
                "dataDigestType": "digestProhibited",
                "dataDigestInherited": true
              },
              "supportedAdvancedOptions": [
                {
                  "label": "ErrorRecoveryLevel",
                  "summary": "iSCSI option : iSCSI Error Recovery Level (ERL) value that the ESX initiator would negotiate during login.",
                  "key": "ErrorRecoveryLevel",
                  "optionType": {
                    "valueIsReadonly": true,
                    "min": 0,
                    "max": 2,
                    "defaultValue": 0
                  }
                },
                {
                  "label": "LoginRetryMax",
                  "summary": "iSCSI option : Maximum number of times ESX initiator would retry login to a target, before giving up.",
                  "key": "LoginRetryMax",
                  "optionType": {
                    "valueIsReadonly": true,
                    "min": 0,
                    "max": 64,
                    "defaultValue": 4
                  }
                },
                {
                  "label": "MaxOutstandingR2T",
                  "summary": "iSCSI option : Maximum number of R2T (Ready To Transfer) PDUs, that can be outstanding for a task.",
                  "key": "MaxOutstandingR2T",
                  "optionType": {
                    "valueIsReadonly": false,
                    "min": 1,
                    "max": 8,
                    "defaultValue": 1
                  }
                },
                {
                  "label": "FirstBurstLength",
                  "summary": "iSCSI option : Maximum unsolicited data in bytes initiator can send during the execution of a single SCSI command. It must be <= MaxBurstLength.",
                  "key": "FirstBurstLength",
                  "optionType": {
                    "valueIsReadonly": false,
                    "min": 512,
                    "max": 16777215,
                    "defaultValue": 262144
                  }
                },
                {
                  "label": "MaxBurstLength",
                  "summary": "iSCSI option : Maximum SCSI data payload in bytes in a Data-In or a solicited Data-Out iSCSI sequence.",
                  "key": "MaxBurstLength",
                  "optionType": {
                    "valueIsReadonly": false,
                    "min": 512,
                    "max": 16777215,
                    "defaultValue": 262144
                  }
                },
                {
                  "label": "MaxRecvDataSegLen",
                  "summary": "iSCSI option : Maximum data segment length in bytes that can be received in an iSCSI PDU. It is recommended to keep it <= MaxBurstLength.",
                  "key": "MaxRecvDataSegLen",
                  "optionType": {
                    "valueIsReadonly": false,
                    "min": 512,
                    "max": 16777215,
                    "defaultValue": 131072
                  }
                },
                {
                  "label": "MaxCommands",
                  "summary": "iSCSI option : Maximum SCSI commands that can be queued on the iscsi adpater.",
                  "key": "MaxCommands",
                  "optionType": {
                    "valueIsReadonly": true,
                    "min": 2,
                    "max": 2048,
                    "defaultValue": 128
                  }
                },
                {
                  "label": "DefaultTimeToWait",
                  "summary": "iSCSI option : Minimum seconds to wait before attempting a logout or an active task reassignment after an unexpected connection termination or reset.",
                  "key": "DefaultTimeToWait",
                  "optionType": {
                    "valueIsReadonly": true,
                    "min": 0,
                    "max": 60,
                    "defaultValue": 2
                  }
                },
                {
                  "label": "DefaultTimeToRetain",
                  "summary": "iSCSI option : Maximum seconds that a connection and task allegiance reinstatement is still possible following a connection termination or reset.",
                  "key": "DefaultTimeToRetain",
                  "optionType": {
                    "valueIsReadonly": true,
                    "min": 0,
                    "max": 60,
                    "defaultValue": 0
                  }
                },
                {
                  "label": "LoginTimeout",
                  "summary": "iSCSI option : Time in seconds initiator will wait for the Login response to finish.",
                  "key": "LoginTimeout",
                  "optionType": {
                    "valueIsReadonly": false,
                    "min": 1,
                    "max": 60,
                    "defaultValue": 5
                  }
                },
                {
                  "label": "LogoutTimeout",
                  "summary": "iSCSI option : Time in seconds initiator will wait to get a response for Logout request PDU",
                  "key": "LogoutTimeout",
                  "optionType": {
                    "valueIsReadonly": true,
                    "min": 0,
                    "max": 60,
                    "defaultValue": 15
                  }
                },
                {
                  "label": "RecoveryTimeout",
                  "summary": "iSCSI option : Time in seconds that can elapse while a session recovery is performed. If the session recovery is not over within this time, initiator terminates the session.",
                  "key": "RecoveryTimeout",
                  "optionType": {
                    "valueIsReadonly": false,
                    "min": 1,
                    "max": 120,
                    "defaultValue": 10
                  }
                },
                {
                  "label": "NoopTimeout",
                  "summary": "iSCSI option : Time in seconds that can elapse before initiator receives a NOP-IN PDU from the target. When no-op timeout limit is exceeded, initiator terminates the current connection and starts a new one.",
                  "key": "NoopTimeout",
                  "optionType": {
                    "valueIsReadonly": false,
                    "min": 10,
                    "max": 30,
                    "defaultValue": 10
                  }
                },
                {
                  "label": "NoopInterval",
                  "summary": "iSCSI option : Time in seconds in between NOP-OUTs sent by the initiator to verify that a connection is still active.",
                  "key": "NoopInterval",
                  "optionType": {
                    "valueIsReadonly": false,
                    "min": 1,
                    "max": 60,
                    "defaultValue": 15
                  }
                },
                {
                  "label": "InitR2T",
                  "summary": "iSCSI option : Whether to allow initiator to start sending data to a target as if it had received an initial R2T.",
                  "key": "InitR2T",
                  "optionType": {
                    "valueIsReadonly": true,
                    "supported": true,
                    "defaultValue": false
                  }
                },
                {
                  "label": "ImmediateData",
                  "summary": "iSCSI option : Whether to allow initiator to send immediate (unsolicited) data to the target.",
                  "key": "ImmediateData",
                  "optionType": {
                    "valueIsReadonly": true,
                    "supported": true,
                    "defaultValue": true
                  }
                },
                {
                  "label": "DelayedAck",
                  "summary": "iSCSI option : Whether to allow initiator to delay acknowledgement of received data packets.",
                  "key": "DelayedAck",
                  "optionType": {
                    "valueIsReadonly": false,
                    "supported": true,
                    "defaultValue": true
                  }
                }
              ],
              "advancedOptions": [
                {
                  "key": "ErrorRecoveryLevel",
                  "value": 0,
                  "isInherited": true
                },
                {
                  "key": "LoginRetryMax",
                  "value": 4,
                  "isInherited": true
                },
                {
                  "key": "MaxOutstandingR2T",
                  "value": 1,
                  "isInherited": true
                },
                {
                  "key": "FirstBurstLength",
                  "value": 262144,
                  "isInherited": true
                },
                {
                  "key": "MaxBurstLength",
                  "value": 262144,
                  "isInherited": true
                },
                {
                  "key": "MaxRecvDataSegLen",
                  "value": 131072,
                  "isInherited": true
                },
                {
                  "key": "MaxCommands",
                  "value": 128,
                  "isInherited": true
                },
                {
                  "key": "DefaultTimeToWait",
                  "value": 0,
                  "isInherited": true
                },
                {
                  "key": "DefaultTimeToRetain",
                  "value": 0,
                  "isInherited": true
                },
                {
                  "key": "LoginTimeout",
                  "value": 5,
                  "isInherited": true
                },
                {
                  "key": "LogoutTimeout",
                  "value": 15,
                  "isInherited": true
                },
                {
                  "key": "RecoveryTimeout",
                  "value": 10,
                  "isInherited": true
                },
                {
                  "key": "NoopTimeout",
                  "value": 10,
                  "isInherited": true
                },
                {
                  "key": "NoopInterval",
                  "value": 15,
                  "isInherited": true
                },
                {
                  "key": "InitR2T",
                  "value": false,
                  "isInherited": true
                },
                {
                  "key": "ImmediateData",
                  "value": false,
                  "isInherited": true
                },
                {
                  "key": "DelayedAck",
                  "value": true,
                  "isInherited": true
                }
              ],
              "parent": "vmhba64"
            },
            {
              "address": "10.136.15.232",
              "port": 3260,
              "authenticationProperties": {
                "chapAuthEnabled": false,
                "chapAuthenticationType": "chapProhibited",
                "chapInherited": true,
                "mutualChapAuthenticationType": "chapProhibited",
                "mutualChapInherited": true
              },
              "digestProperties": {
                "headerDigestType": "digestProhibited",
                "headerDigestInherited": true,
                "dataDigestType": "digestProhibited",
                "dataDigestInherited": true
              },
              "supportedAdvancedOptions": [
                {
                  "label": "ErrorRecoveryLevel",
                  "summary": "iSCSI option : iSCSI Error Recovery Level (ERL) value that the ESX initiator would negotiate during login.",
                  "key": "ErrorRecoveryLevel",
                  "optionType": {
                    "valueIsReadonly": true,
                    "min": 0,
                    "max": 2,
                    "defaultValue": 0
                  }
                },
                {
                  "label": "LoginRetryMax",
                  "summary": "iSCSI option : Maximum number of times ESX initiator would retry login to a target, before giving up.",
                  "key": "LoginRetryMax",
                  "optionType": {
                    "valueIsReadonly": true,
                    "min": 0,
                    "max": 64,
                    "defaultValue": 4
                  }
                },
                {
                  "label": "MaxOutstandingR2T",
                  "summary": "iSCSI option : Maximum number of R2T (Ready To Transfer) PDUs, that can be outstanding for a task.",
                  "key": "MaxOutstandingR2T",
                  "optionType": {
                    "valueIsReadonly": false,
                    "min": 1,
                    "max": 8,
                    "defaultValue": 1
                  }
                },
                {
                  "label": "FirstBurstLength",
                  "summary": "iSCSI option : Maximum unsolicited data in bytes initiator can send during the execution of a single SCSI command. It must be <= MaxBurstLength.",
                  "key": "FirstBurstLength",
                  "optionType": {
                    "valueIsReadonly": false,
                    "min": 512,
                    "max": 16777215,
                    "defaultValue": 262144
                  }
                },
                {
                  "label": "MaxBurstLength",
                  "summary": "iSCSI option : Maximum SCSI data payload in bytes in a Data-In or a solicited Data-Out iSCSI sequence.",
                  "key": "MaxBurstLength",
                  "optionType": {
                    "valueIsReadonly": false,
                    "min": 512,
                    "max": 16777215,
                    "defaultValue": 262144
                  }
                },
                {
                  "label": "MaxRecvDataSegLen",
                  "summary": "iSCSI option : Maximum data segment length in bytes that can be received in an iSCSI PDU. It is recommended to keep it <= MaxBurstLength.",
                  "key": "MaxRecvDataSegLen",
                  "optionType": {
                    "valueIsReadonly": false,
                    "min": 512,
                    "max": 16777215,
                    "defaultValue": 131072
                  }
                },
                {
                  "label": "MaxCommands",
                  "summary": "iSCSI option : Maximum SCSI commands that can be queued on the iscsi adpater.",
                  "key": "MaxCommands",
                  "optionType": {
                    "valueIsReadonly": true,
                    "min": 2,
                    "max": 2048,
                    "defaultValue": 128
                  }
                },
                {
                  "label": "DefaultTimeToWait",
                  "summary": "iSCSI option : Minimum seconds to wait before attempting a logout or an active task reassignment after an unexpected connection termination or reset.",
                  "key": "DefaultTimeToWait",
                  "optionType": {
                    "valueIsReadonly": true,
                    "min": 0,
                    "max": 60,
                    "defaultValue": 2
                  }
                },
                {
                  "label": "DefaultTimeToRetain",
                  "summary": "iSCSI option : Maximum seconds that a connection and task allegiance reinstatement is still possible following a connection termination or reset.",
                  "key": "DefaultTimeToRetain",
                  "optionType": {
                    "valueIsReadonly": true,
                    "min": 0,
                    "max": 60,
                    "defaultValue": 0
                  }
                },
                {
                  "label": "LoginTimeout",
                  "summary": "iSCSI option : Time in seconds initiator will wait for the Login response to finish.",
                  "key": "LoginTimeout",
                  "optionType": {
                    "valueIsReadonly": false,
                    "min": 1,
                    "max": 60,
                    "defaultValue": 5
                  }
                },
                {
                  "label": "LogoutTimeout",
                  "summary": "iSCSI option : Time in seconds initiator will wait to get a response for Logout request PDU",
                  "key": "LogoutTimeout",
                  "optionType": {
                    "valueIsReadonly": true,
                    "min": 0,
                    "max": 60,
                    "defaultValue": 15
                  }
                },
                {
                  "label": "RecoveryTimeout",
                  "summary": "iSCSI option : Time in seconds that can elapse while a session recovery is performed. If the session recovery is not over within this time, initiator terminates the session.",
                  "key": "RecoveryTimeout",
                  "optionType": {
                    "valueIsReadonly": false,
                    "min": 1,
                    "max": 120,
                    "defaultValue": 10
                  }
                },
                {
                  "label": "NoopTimeout",
                  "summary": "iSCSI option : Time in seconds that can elapse before initiator receives a NOP-IN PDU from the target. When no-op timeout limit is exceeded, initiator terminates the current connection and starts a new one.",
                  "key": "NoopTimeout",
                  "optionType": {
                    "valueIsReadonly": false,
                    "min": 10,
                    "max": 30,
                    "defaultValue": 10
                  }
                },
                {
                  "label": "NoopInterval",
                  "summary": "iSCSI option : Time in seconds in between NOP-OUTs sent by the initiator to verify that a connection is still active.",
                  "key": "NoopInterval",
                  "optionType": {
                    "valueIsReadonly": false,
                    "min": 1,
                    "max": 60,
                    "defaultValue": 15
                  }
                },
                {
                  "label": "InitR2T",
                  "summary": "iSCSI option : Whether to allow initiator to start sending data to a target as if it had received an initial R2T.",
                  "key": "InitR2T",
                  "optionType": {
                    "valueIsReadonly": true,
                    "supported": true,
                    "defaultValue": false
                  }
                },
                {
                  "label": "ImmediateData",
                  "summary": "iSCSI option : Whether to allow initiator to send immediate (unsolicited) data to the target.",
                  "key": "ImmediateData",
                  "optionType": {
                    "valueIsReadonly": true,
                    "supported": true,
                    "defaultValue": true
                  }
                },
                {
                  "label": "DelayedAck",
                  "summary": "iSCSI option : Whether to allow initiator to delay acknowledgement of received data packets.",
                  "key": "DelayedAck",
                  "optionType": {
                    "valueIsReadonly": false,
                    "supported": true,
                    "defaultValue": true
                  }
                }
              ],
              "advancedOptions": [
                {
                  "key": "ErrorRecoveryLevel",
                  "value": 0,
                  "isInherited": true
                },
                {
                  "key": "LoginRetryMax",
                  "value": 4,
                  "isInherited": true
                },
                {
                  "key": "MaxOutstandingR2T",
                  "value": 1,
                  "isInherited": true
                },
                {
                  "key": "FirstBurstLength",
                  "value": 262144,
                  "isInherited": true
                },
                {
                  "key": "MaxBurstLength",
                  "value": 262144,
                  "isInherited": true
                },
                {
                  "key": "MaxRecvDataSegLen",
                  "value": 131072,
                  "isInherited": true
                },
                {
                  "key": "MaxCommands",
                  "value": 128,
                  "isInherited": true
                },
                {
                  "key": "DefaultTimeToWait",
                  "value": 0,
                  "isInherited": true
                },
                {
                  "key": "DefaultTimeToRetain",
                  "value": 0,
                  "isInherited": true
                },
                {
                  "key": "LoginTimeout",
                  "value": 5,
                  "isInherited": true
                },
                {
                  "key": "LogoutTimeout",
                  "value": 15,
                  "isInherited": true
                },
                {
                  "key": "RecoveryTimeout",
                  "value": 10,
                  "isInherited": true
                },
                {
                  "key": "NoopTimeout",
                  "value": 10,
                  "isInherited": true
                },
                {
                  "key": "NoopInterval",
                  "value": 15,
                  "isInherited": true
                },
                {
                  "key": "InitR2T",
                  "value": false,
                  "isInherited": true
                },
                {
                  "key": "ImmediateData",
                  "value": false,
                  "isInherited": true
                },
                {
                  "key": "DelayedAck",
                  "value": true,
                  "isInherited": true
                }
              ],
              "parent": "vmhba64"
            }
          ]"""
static_targets = """[
            {
              "address": "10.136.13.230",
              "port": 3260,
              "iScsiName": "iqn.2010-06.com.purestorage:flasharray.15dbd8b7f06a7c64",
              "discoveryMethod": "sendTargetMethod",
              "authenticationProperties": {
                "chapAuthEnabled": false,
                "chapAuthenticationType": "chapProhibited",
                "chapInherited": true,
                "mutualChapAuthenticationType": "chapProhibited",
                "mutualChapInherited": true
              },
              "digestProperties": {
                "headerDigestType": "digestProhibited",
                "headerDigestInherited": true,
                "dataDigestType": "digestProhibited",
                "dataDigestInherited": true
              },
              "supportedAdvancedOptions": [
                {
                  "label": "ErrorRecoveryLevel",
                  "summary": "iSCSI option : iSCSI Error Recovery Level (ERL) value that the ESX initiator would negotiate during login.",
                  "key": "ErrorRecoveryLevel",
                  "optionType": {
                    "valueIsReadonly": true,
                    "min": 0,
                    "max": 2,
                    "defaultValue": 0
                  }
                },
                {
                  "label": "LoginRetryMax",
                  "summary": "iSCSI option : Maximum number of times ESX initiator would retry login to a target, before giving up.",
                  "key": "LoginRetryMax",
                  "optionType": {
                    "valueIsReadonly": true,
                    "min": 0,
                    "max": 64,
                    "defaultValue": 4
                  }
                },
                {
                  "label": "MaxOutstandingR2T",
                  "summary": "iSCSI option : Maximum number of R2T (Ready To Transfer) PDUs, that can be outstanding for a task.",
                  "key": "MaxOutstandingR2T",
                  "optionType": {
                    "valueIsReadonly": false,
                    "min": 1,
                    "max": 8,
                    "defaultValue": 1
                  }
                },
                {
                  "label": "FirstBurstLength",
                  "summary": "iSCSI option : Maximum unsolicited data in bytes initiator can send during the execution of a single SCSI command. It must be <= MaxBurstLength.",
                  "key": "FirstBurstLength",
                  "optionType": {
                    "valueIsReadonly": false,
                    "min": 512,
                    "max": 16777215,
                    "defaultValue": 262144
                  }
                },
                {
                  "label": "MaxBurstLength",
                  "summary": "iSCSI option : Maximum SCSI data payload in bytes in a Data-In or a solicited Data-Out iSCSI sequence.",
                  "key": "MaxBurstLength",
                  "optionType": {
                    "valueIsReadonly": false,
                    "min": 512,
                    "max": 16777215,
                    "defaultValue": 262144
                  }
                },
                {
                  "label": "MaxRecvDataSegLen",
                  "summary": "iSCSI option : Maximum data segment length in bytes that can be received in an iSCSI PDU. It is recommended to keep it <= MaxBurstLength.",
                  "key": "MaxRecvDataSegLen",
                  "optionType": {
                    "valueIsReadonly": false,
                    "min": 512,
                    "max": 16777215,
                    "defaultValue": 131072
                  }
                },
                {
                  "label": "MaxCommands",
                  "summary": "iSCSI option : Maximum SCSI commands that can be queued on the iscsi adpater.",
                  "key": "MaxCommands",
                  "optionType": {
                    "valueIsReadonly": true,
                    "min": 2,
                    "max": 2048,
                    "defaultValue": 128
                  }
                },
                {
                  "label": "DefaultTimeToWait",
                  "summary": "iSCSI option : Minimum seconds to wait before attempting a logout or an active task reassignment after an unexpected connection termination or reset.",
                  "key": "DefaultTimeToWait",
                  "optionType": {
                    "valueIsReadonly": true,
                    "min": 0,
                    "max": 60,
                    "defaultValue": 2
                  }
                },
                {
                  "label": "DefaultTimeToRetain",
                  "summary": "iSCSI option : Maximum seconds that a connection and task allegiance reinstatement is still possible following a connection termination or reset.",
                  "key": "DefaultTimeToRetain",
                  "optionType": {
                    "valueIsReadonly": true,
                    "min": 0,
                    "max": 60,
                    "defaultValue": 0
                  }
                },
                {
                  "label": "LoginTimeout",
                  "summary": "iSCSI option : Time in seconds initiator will wait for the Login response to finish.",
                  "key": "LoginTimeout",
                  "optionType": {
                    "valueIsReadonly": false,
                    "min": 1,
                    "max": 60,
                    "defaultValue": 5
                  }
                },
                {
                  "label": "LogoutTimeout",
                  "summary": "iSCSI option : Time in seconds initiator will wait to get a response for Logout request PDU",
                  "key": "LogoutTimeout",
                  "optionType": {
                    "valueIsReadonly": true,
                    "min": 0,
                    "max": 60,
                    "defaultValue": 15
                  }
                },
                {
                  "label": "RecoveryTimeout",
                  "summary": "iSCSI option : Time in seconds that can elapse while a session recovery is performed. If the session recovery is not over within this time, initiator terminates the session.",
                  "key": "RecoveryTimeout",
                  "optionType": {
                    "valueIsReadonly": false,
                    "min": 1,
                    "max": 120,
                    "defaultValue": 10
                  }
                },
                {
                  "label": "NoopTimeout",
                  "summary": "iSCSI option : Time in seconds that can elapse before initiator receives a NOP-IN PDU from the target. When no-op timeout limit is exceeded, initiator terminates the current connection and starts a new one.",
                  "key": "NoopTimeout",
                  "optionType": {
                    "valueIsReadonly": false,
                    "min": 10,
                    "max": 30,
                    "defaultValue": 10
                  }
                },
                {
                  "label": "NoopInterval",
                  "summary": "iSCSI option : Time in seconds in between NOP-OUTs sent by the initiator to verify that a connection is still active.",
                  "key": "NoopInterval",
                  "optionType": {
                    "valueIsReadonly": false,
                    "min": 1,
                    "max": 60,
                    "defaultValue": 15
                  }
                },
                {
                  "label": "InitR2T",
                  "summary": "iSCSI option : Whether to allow initiator to start sending data to a target as if it had received an initial R2T.",
                  "key": "InitR2T",
                  "optionType": {
                    "valueIsReadonly": true,
                    "supported": true,
                    "defaultValue": false
                  }
                },
                {
                  "label": "ImmediateData",
                  "summary": "iSCSI option : Whether to allow initiator to send immediate (unsolicited) data to the target.",
                  "key": "ImmediateData",
                  "optionType": {
                    "valueIsReadonly": true,
                    "supported": true,
                    "defaultValue": true
                  }
                },
                {
                  "label": "DelayedAck",
                  "summary": "iSCSI option : Whether to allow initiator to delay acknowledgement of received data packets.",
                  "key": "DelayedAck",
                  "optionType": {
                    "valueIsReadonly": false,
                    "supported": true,
                    "defaultValue": true
                  }
                }
              ],
              "advancedOptions": [
                {
                  "key": "ErrorRecoveryLevel",
                  "value": 0,
                  "isInherited": true
                },
                {
                  "key": "LoginRetryMax",
                  "value": 4,
                  "isInherited": true
                },
                {
                  "key": "MaxOutstandingR2T",
                  "value": 1,
                  "isInherited": true
                },
                {
                  "key": "FirstBurstLength",
                  "value": 262144,
                  "isInherited": true
                },
                {
                  "key": "MaxBurstLength",
                  "value": 262144,
                  "isInherited": true
                },
                {
                  "key": "MaxRecvDataSegLen",
                  "value": 131072,
                  "isInherited": true
                },
                {
                  "key": "MaxCommands",
                  "value": 128,
                  "isInherited": true
                },
                {
                  "key": "DefaultTimeToWait",
                  "value": 2,
                  "isInherited": true
                },
                {
                  "key": "DefaultTimeToRetain",
                  "value": 0,
                  "isInherited": true
                },
                {
                  "key": "LoginTimeout",
                  "value": 5,
                  "isInherited": true
                },
                {
                  "key": "LogoutTimeout",
                  "value": 15,
                  "isInherited": true
                },
                {
                  "key": "RecoveryTimeout",
                  "value": 10,
                  "isInherited": true
                },
                {
                  "key": "NoopTimeout",
                  "value": 10,
                  "isInherited": true
                },
                {
                  "key": "NoopInterval",
                  "value": 15,
                  "isInherited": true
                },
                {
                  "key": "InitR2T",
                  "value": false,
                  "isInherited": true
                },
                {
                  "key": "ImmediateData",
                  "value": true,
                  "isInherited": true
                },
                {
                  "key": "DelayedAck",
                  "value": true,
                  "isInherited": true
                }
              ],
              "parent": "10.136.15.233:3260"
            },
            {
              "address": "10.136.13.233",
              "port": 3260,
              "iScsiName": "iqn.2010-06.com.purestorage:flasharray.15dbd8b7f06a7c64",
              "discoveryMethod": "sendTargetMethod",
              "authenticationProperties": {
                "chapAuthEnabled": false,
                "chapAuthenticationType": "chapProhibited",
                "chapInherited": true,
                "mutualChapAuthenticationType": "chapProhibited",
                "mutualChapInherited": true
              },
              "digestProperties": {
                "headerDigestType": "digestProhibited",
                "headerDigestInherited": true,
                "dataDigestType": "digestProhibited",
                "dataDigestInherited": true
              },
              "supportedAdvancedOptions": [
                {
                  "label": "ErrorRecoveryLevel",
                  "summary": "iSCSI option : iSCSI Error Recovery Level (ERL) value that the ESX initiator would negotiate during login.",
                  "key": "ErrorRecoveryLevel",
                  "optionType": {
                    "valueIsReadonly": true,
                    "min": 0,
                    "max": 2,
                    "defaultValue": 0
                  }
                },
                {
                  "label": "LoginRetryMax",
                  "summary": "iSCSI option : Maximum number of times ESX initiator would retry login to a target, before giving up.",
                  "key": "LoginRetryMax",
                  "optionType": {
                    "valueIsReadonly": true,
                    "min": 0,
                    "max": 64,
                    "defaultValue": 4
                  }
                },
                {
                  "label": "MaxOutstandingR2T",
                  "summary": "iSCSI option : Maximum number of R2T (Ready To Transfer) PDUs, that can be outstanding for a task.",
                  "key": "MaxOutstandingR2T",
                  "optionType": {
                    "valueIsReadonly": false,
                    "min": 1,
                    "max": 8,
                    "defaultValue": 1
                  }
                },
                {
                  "label": "FirstBurstLength",
                  "summary": "iSCSI option : Maximum unsolicited data in bytes initiator can send during the execution of a single SCSI command. It must be <= MaxBurstLength.",
                  "key": "FirstBurstLength",
                  "optionType": {
                    "valueIsReadonly": false,
                    "min": 512,
                    "max": 16777215,
                    "defaultValue": 262144
                  }
                },
                {
                  "label": "MaxBurstLength",
                  "summary": "iSCSI option : Maximum SCSI data payload in bytes in a Data-In or a solicited Data-Out iSCSI sequence.",
                  "key": "MaxBurstLength",
                  "optionType": {
                    "valueIsReadonly": false,
                    "min": 512,
                    "max": 16777215,
                    "defaultValue": 262144
                  }
                },
                {
                  "label": "MaxRecvDataSegLen",
                  "summary": "iSCSI option : Maximum data segment length in bytes that can be received in an iSCSI PDU. It is recommended to keep it <= MaxBurstLength.",
                  "key": "MaxRecvDataSegLen",
                  "optionType": {
                    "valueIsReadonly": false,
                    "min": 512,
                    "max": 16777215,
                    "defaultValue": 131072
                  }
                },
                {
                  "label": "MaxCommands",
                  "summary": "iSCSI option : Maximum SCSI commands that can be queued on the iscsi adpater.",
                  "key": "MaxCommands",
                  "optionType": {
                    "valueIsReadonly": true,
                    "min": 2,
                    "max": 2048,
                    "defaultValue": 128
                  }
                },
                {
                  "label": "DefaultTimeToWait",
                  "summary": "iSCSI option : Minimum seconds to wait before attempting a logout or an active task reassignment after an unexpected connection termination or reset.",
                  "key": "DefaultTimeToWait",
                  "optionType": {
                    "valueIsReadonly": true,
                    "min": 0,
                    "max": 60,
                    "defaultValue": 2
                  }
                },
                {
                  "label": "DefaultTimeToRetain",
                  "summary": "iSCSI option : Maximum seconds that a connection and task allegiance reinstatement is still possible following a connection termination or reset.",
                  "key": "DefaultTimeToRetain",
                  "optionType": {
                    "valueIsReadonly": true,
                    "min": 0,
                    "max": 60,
                    "defaultValue": 0
                  }
                },
                {
                  "label": "LoginTimeout",
                  "summary": "iSCSI option : Time in seconds initiator will wait for the Login response to finish.",
                  "key": "LoginTimeout",
                  "optionType": {
                    "valueIsReadonly": false,
                    "min": 1,
                    "max": 60,
                    "defaultValue": 5
                  }
                },
                {
                  "label": "LogoutTimeout",
                  "summary": "iSCSI option : Time in seconds initiator will wait to get a response for Logout request PDU",
                  "key": "LogoutTimeout",
                  "optionType": {
                    "valueIsReadonly": true,
                    "min": 0,
                    "max": 60,
                    "defaultValue": 15
                  }
                },
                {
                  "label": "RecoveryTimeout",
                  "summary": "iSCSI option : Time in seconds that can elapse while a session recovery is performed. If the session recovery is not over within this time, initiator terminates the session.",
                  "key": "RecoveryTimeout",
                  "optionType": {
                    "valueIsReadonly": false,
                    "min": 1,
                    "max": 120,
                    "defaultValue": 10
                  }
                },
                {
                  "label": "NoopTimeout",
                  "summary": "iSCSI option : Time in seconds that can elapse before initiator receives a NOP-IN PDU from the target. When no-op timeout limit is exceeded, initiator terminates the current connection and starts a new one.",
                  "key": "NoopTimeout",
                  "optionType": {
                    "valueIsReadonly": false,
                    "min": 10,
                    "max": 30,
                    "defaultValue": 10
                  }
                },
                {
                  "label": "NoopInterval",
                  "summary": "iSCSI option : Time in seconds in between NOP-OUTs sent by the initiator to verify that a connection is still active.",
                  "key": "NoopInterval",
                  "optionType": {
                    "valueIsReadonly": false,
                    "min": 1,
                    "max": 60,
                    "defaultValue": 15
                  }
                },
                {
                  "label": "InitR2T",
                  "summary": "iSCSI option : Whether to allow initiator to start sending data to a target as if it had received an initial R2T.",
                  "key": "InitR2T",
                  "optionType": {
                    "valueIsReadonly": true,
                    "supported": true,
                    "defaultValue": false
                  }
                },
                {
                  "label": "ImmediateData",
                  "summary": "iSCSI option : Whether to allow initiator to send immediate (unsolicited) data to the target.",
                  "key": "ImmediateData",
                  "optionType": {
                    "valueIsReadonly": true,
                    "supported": true,
                    "defaultValue": true
                  }
                },
                {
                  "label": "DelayedAck",
                  "summary": "iSCSI option : Whether to allow initiator to delay acknowledgement of received data packets.",
                  "key": "DelayedAck",
                  "optionType": {
                    "valueIsReadonly": false,
                    "supported": true,
                    "defaultValue": true
                  }
                }
              ],
              "advancedOptions": [
                {
                  "key": "ErrorRecoveryLevel",
                  "value": 0,
                  "isInherited": true
                },
                {
                  "key": "LoginRetryMax",
                  "value": 4,
                  "isInherited": true
                },
                {
                  "key": "MaxOutstandingR2T",
                  "value": 1,
                  "isInherited": true
                },
                {
                  "key": "FirstBurstLength",
                  "value": 262144,
                  "isInherited": true
                },
                {
                  "key": "MaxBurstLength",
                  "value": 262144,
                  "isInherited": true
                },
                {
                  "key": "MaxRecvDataSegLen",
                  "value": 131072,
                  "isInherited": true
                },
                {
                  "key": "MaxCommands",
                  "value": 128,
                  "isInherited": true
                },
                {
                  "key": "DefaultTimeToWait",
                  "value": 2,
                  "isInherited": true
                },
                {
                  "key": "DefaultTimeToRetain",
                  "value": 0,
                  "isInherited": true
                },
                {
                  "key": "LoginTimeout",
                  "value": 5,
                  "isInherited": true
                },
                {
                  "key": "LogoutTimeout",
                  "value": 15,
                  "isInherited": true
                },
                {
                  "key": "RecoveryTimeout",
                  "value": 10,
                  "isInherited": true
                },
                {
                  "key": "NoopTimeout",
                  "value": 10,
                  "isInherited": true
                },
                {
                  "key": "NoopInterval",
                  "value": 15,
                  "isInherited": true
                },
                {
                  "key": "InitR2T",
                  "value": false,
                  "isInherited": true
                },
                {
                  "key": "ImmediateData",
                  "value": true,
                  "isInherited": true
                },
                {
                  "key": "DelayedAck",
                  "value": true,
                  "isInherited": true
                }
              ],
              "parent": "10.136.15.233:3260"
            },
            {
              "address": "10.136.15.233",
              "port": 3260,
              "iScsiName": "iqn.2010-06.com.purestorage:flasharray.15dbd8b7f06a7c64",
              "discoveryMethod": "sendTargetMethod",
              "authenticationProperties": {
                "chapAuthEnabled": false,
                "chapAuthenticationType": "chapProhibited",
                "chapInherited": true,
                "mutualChapAuthenticationType": "chapProhibited",
                "mutualChapInherited": true
              },
              "digestProperties": {
                "headerDigestType": "digestProhibited",
                "headerDigestInherited": true,
                "dataDigestType": "digestProhibited",
                "dataDigestInherited": true
              },
              "supportedAdvancedOptions": [
                {
                  "label": "ErrorRecoveryLevel",
                  "summary": "iSCSI option : iSCSI Error Recovery Level (ERL) value that the ESX initiator would negotiate during login.",
                  "key": "ErrorRecoveryLevel",
                  "optionType": {
                    "valueIsReadonly": true,
                    "min": 0,
                    "max": 2,
                    "defaultValue": 0
                  }
                },
                {
                  "label": "LoginRetryMax",
                  "summary": "iSCSI option : Maximum number of times ESX initiator would retry login to a target, before giving up.",
                  "key": "LoginRetryMax",
                  "optionType": {
                    "valueIsReadonly": true,
                    "min": 0,
                    "max": 64,
                    "defaultValue": 4
                  }
                },
                {
                  "label": "MaxOutstandingR2T",
                  "summary": "iSCSI option : Maximum number of R2T (Ready To Transfer) PDUs, that can be outstanding for a task.",
                  "key": "MaxOutstandingR2T",
                  "optionType": {
                    "valueIsReadonly": false,
                    "min": 1,
                    "max": 8,
                    "defaultValue": 1
                  }
                },
                {
                  "label": "FirstBurstLength",
                  "summary": "iSCSI option : Maximum unsolicited data in bytes initiator can send during the execution of a single SCSI command. It must be <= MaxBurstLength.",
                  "key": "FirstBurstLength",
                  "optionType": {
                    "valueIsReadonly": false,
                    "min": 512,
                    "max": 16777215,
                    "defaultValue": 262144
                  }
                },
                {
                  "label": "MaxBurstLength",
                  "summary": "iSCSI option : Maximum SCSI data payload in bytes in a Data-In or a solicited Data-Out iSCSI sequence.",
                  "key": "MaxBurstLength",
                  "optionType": {
                    "valueIsReadonly": false,
                    "min": 512,
                    "max": 16777215,
                    "defaultValue": 262144
                  }
                },
                {
                  "label": "MaxRecvDataSegLen",
                  "summary": "iSCSI option : Maximum data segment length in bytes that can be received in an iSCSI PDU. It is recommended to keep it <= MaxBurstLength.",
                  "key": "MaxRecvDataSegLen",
                  "optionType": {
                    "valueIsReadonly": false,
                    "min": 512,
                    "max": 16777215,
                    "defaultValue": 131072
                  }
                },
                {
                  "label": "MaxCommands",
                  "summary": "iSCSI option : Maximum SCSI commands that can be queued on the iscsi adpater.",
                  "key": "MaxCommands",
                  "optionType": {
                    "valueIsReadonly": true,
                    "min": 2,
                    "max": 2048,
                    "defaultValue": 128
                  }
                },
                {
                  "label": "DefaultTimeToWait",
                  "summary": "iSCSI option : Minimum seconds to wait before attempting a logout or an active task reassignment after an unexpected connection termination or reset.",
                  "key": "DefaultTimeToWait",
                  "optionType": {
                    "valueIsReadonly": true,
                    "min": 0,
                    "max": 60,
                    "defaultValue": 2
                  }
                },
                {
                  "label": "DefaultTimeToRetain",
                  "summary": "iSCSI option : Maximum seconds that a connection and task allegiance reinstatement is still possible following a connection termination or reset.",
                  "key": "DefaultTimeToRetain",
                  "optionType": {
                    "valueIsReadonly": true,
                    "min": 0,
                    "max": 60,
                    "defaultValue": 0
                  }
                },
                {
                  "label": "LoginTimeout",
                  "summary": "iSCSI option : Time in seconds initiator will wait for the Login response to finish.",
                  "key": "LoginTimeout",
                  "optionType": {
                    "valueIsReadonly": false,
                    "min": 1,
                    "max": 60,
                    "defaultValue": 5
                  }
                },
                {
                  "label": "LogoutTimeout",
                  "summary": "iSCSI option : Time in seconds initiator will wait to get a response for Logout request PDU",
                  "key": "LogoutTimeout",
                  "optionType": {
                    "valueIsReadonly": true,
                    "min": 0,
                    "max": 60,
                    "defaultValue": 15
                  }
                },
                {
                  "label": "RecoveryTimeout",
                  "summary": "iSCSI option : Time in seconds that can elapse while a session recovery is performed. If the session recovery is not over within this time, initiator terminates the session.",
                  "key": "RecoveryTimeout",
                  "optionType": {
                    "valueIsReadonly": false,
                    "min": 1,
                    "max": 120,
                    "defaultValue": 10
                  }
                },
                {
                  "label": "NoopTimeout",
                  "summary": "iSCSI option : Time in seconds that can elapse before initiator receives a NOP-IN PDU from the target. When no-op timeout limit is exceeded, initiator terminates the current connection and starts a new one.",
                  "key": "NoopTimeout",
                  "optionType": {
                    "valueIsReadonly": false,
                    "min": 10,
                    "max": 30,
                    "defaultValue": 10
                  }
                },
                {
                  "label": "NoopInterval",
                  "summary": "iSCSI option : Time in seconds in between NOP-OUTs sent by the initiator to verify that a connection is still active.",
                  "key": "NoopInterval",
                  "optionType": {
                    "valueIsReadonly": false,
                    "min": 1,
                    "max": 60,
                    "defaultValue": 15
                  }
                },
                {
                  "label": "InitR2T",
                  "summary": "iSCSI option : Whether to allow initiator to start sending data to a target as if it had received an initial R2T.",
                  "key": "InitR2T",
                  "optionType": {
                    "valueIsReadonly": true,
                    "supported": true,
                    "defaultValue": false
                  }
                },
                {
                  "label": "ImmediateData",
                  "summary": "iSCSI option : Whether to allow initiator to send immediate (unsolicited) data to the target.",
                  "key": "ImmediateData",
                  "optionType": {
                    "valueIsReadonly": true,
                    "supported": true,
                    "defaultValue": true
                  }
                },
                {
                  "label": "DelayedAck",
                  "summary": "iSCSI option : Whether to allow initiator to delay acknowledgement of received data packets.",
                  "key": "DelayedAck",
                  "optionType": {
                    "valueIsReadonly": false,
                    "supported": true,
                    "defaultValue": true
                  }
                }
              ],
              "advancedOptions": [
                {
                  "key": "ErrorRecoveryLevel",
                  "value": 0,
                  "isInherited": true
                },
                {
                  "key": "LoginRetryMax",
                  "value": 4,
                  "isInherited": true
                },
                {
                  "key": "MaxOutstandingR2T",
                  "value": 1,
                  "isInherited": true
                },
                {
                  "key": "FirstBurstLength",
                  "value": 262144,
                  "isInherited": true
                },
                {
                  "key": "MaxBurstLength",
                  "value": 262144,
                  "isInherited": true
                },
                {
                  "key": "MaxRecvDataSegLen",
                  "value": 131072,
                  "isInherited": true
                },
                {
                  "key": "MaxCommands",
                  "value": 128,
                  "isInherited": true
                },
                {
                  "key": "DefaultTimeToWait",
                  "value": 2,
                  "isInherited": true
                },
                {
                  "key": "DefaultTimeToRetain",
                  "value": 0,
                  "isInherited": true
                },
                {
                  "key": "LoginTimeout",
                  "value": 5,
                  "isInherited": true
                },
                {
                  "key": "LogoutTimeout",
                  "value": 15,
                  "isInherited": true
                },
                {
                  "key": "RecoveryTimeout",
                  "value": 10,
                  "isInherited": true
                },
                {
                  "key": "NoopTimeout",
                  "value": 10,
                  "isInherited": true
                },
                {
                  "key": "NoopInterval",
                  "value": 15,
                  "isInherited": true
                },
                {
                  "key": "InitR2T",
                  "value": false,
                  "isInherited": true
                },
                {
                  "key": "ImmediateData",
                  "value": true,
                  "isInherited": true
                },
                {
                  "key": "DelayedAck",
                  "value": true,
                  "isInherited": true
                }
              ],
              "parent": "10.136.15.233:3260"
            },
            {
              "address": "10.136.13.232",
              "port": 3260,
              "iScsiName": "iqn.2010-06.com.purestorage:flasharray.15dbd8b7f06a7c64",
              "discoveryMethod": "sendTargetMethod",
              "authenticationProperties": {
                "chapAuthEnabled": false,
                "chapAuthenticationType": "chapProhibited",
                "chapInherited": true,
                "mutualChapAuthenticationType": "chapProhibited",
                "mutualChapInherited": true
              },
              "digestProperties": {
                "headerDigestType": "digestProhibited",
                "headerDigestInherited": true,
                "dataDigestType": "digestProhibited",
                "dataDigestInherited": true
              },
              "supportedAdvancedOptions": [
                {
                  "label": "ErrorRecoveryLevel",
                  "summary": "iSCSI option : iSCSI Error Recovery Level (ERL) value that the ESX initiator would negotiate during login.",
                  "key": "ErrorRecoveryLevel",
                  "optionType": {
                    "valueIsReadonly": true,
                    "min": 0,
                    "max": 2,
                    "defaultValue": 0
                  }
                },
                {
                  "label": "LoginRetryMax",
                  "summary": "iSCSI option : Maximum number of times ESX initiator would retry login to a target, before giving up.",
                  "key": "LoginRetryMax",
                  "optionType": {
                    "valueIsReadonly": true,
                    "min": 0,
                    "max": 64,
                    "defaultValue": 4
                  }
                },
                {
                  "label": "MaxOutstandingR2T",
                  "summary": "iSCSI option : Maximum number of R2T (Ready To Transfer) PDUs, that can be outstanding for a task.",
                  "key": "MaxOutstandingR2T",
                  "optionType": {
                    "valueIsReadonly": false,
                    "min": 1,
                    "max": 8,
                    "defaultValue": 1
                  }
                },
                {
                  "label": "FirstBurstLength",
                  "summary": "iSCSI option : Maximum unsolicited data in bytes initiator can send during the execution of a single SCSI command. It must be <= MaxBurstLength.",
                  "key": "FirstBurstLength",
                  "optionType": {
                    "valueIsReadonly": false,
                    "min": 512,
                    "max": 16777215,
                    "defaultValue": 262144
                  }
                },
                {
                  "label": "MaxBurstLength",
                  "summary": "iSCSI option : Maximum SCSI data payload in bytes in a Data-In or a solicited Data-Out iSCSI sequence.",
                  "key": "MaxBurstLength",
                  "optionType": {
                    "valueIsReadonly": false,
                    "min": 512,
                    "max": 16777215,
                    "defaultValue": 262144
                  }
                },
                {
                  "label": "MaxRecvDataSegLen",
                  "summary": "iSCSI option : Maximum data segment length in bytes that can be received in an iSCSI PDU. It is recommended to keep it <= MaxBurstLength.",
                  "key": "MaxRecvDataSegLen",
                  "optionType": {
                    "valueIsReadonly": false,
                    "min": 512,
                    "max": 16777215,
                    "defaultValue": 131072
                  }
                },
                {
                  "label": "MaxCommands",
                  "summary": "iSCSI option : Maximum SCSI commands that can be queued on the iscsi adpater.",
                  "key": "MaxCommands",
                  "optionType": {
                    "valueIsReadonly": true,
                    "min": 2,
                    "max": 2048,
                    "defaultValue": 128
                  }
                },
                {
                  "label": "DefaultTimeToWait",
                  "summary": "iSCSI option : Minimum seconds to wait before attempting a logout or an active task reassignment after an unexpected connection termination or reset.",
                  "key": "DefaultTimeToWait",
                  "optionType": {
                    "valueIsReadonly": true,
                    "min": 0,
                    "max": 60,
                    "defaultValue": 2
                  }
                },
                {
                  "label": "DefaultTimeToRetain",
                  "summary": "iSCSI option : Maximum seconds that a connection and task allegiance reinstatement is still possible following a connection termination or reset.",
                  "key": "DefaultTimeToRetain",
                  "optionType": {
                    "valueIsReadonly": true,
                    "min": 0,
                    "max": 60,
                    "defaultValue": 0
                  }
                },
                {
                  "label": "LoginTimeout",
                  "summary": "iSCSI option : Time in seconds initiator will wait for the Login response to finish.",
                  "key": "LoginTimeout",
                  "optionType": {
                    "valueIsReadonly": false,
                    "min": 1,
                    "max": 60,
                    "defaultValue": 5
                  }
                },
                {
                  "label": "LogoutTimeout",
                  "summary": "iSCSI option : Time in seconds initiator will wait to get a response for Logout request PDU",
                  "key": "LogoutTimeout",
                  "optionType": {
                    "valueIsReadonly": true,
                    "min": 0,
                    "max": 60,
                    "defaultValue": 15
                  }
                },
                {
                  "label": "RecoveryTimeout",
                  "summary": "iSCSI option : Time in seconds that can elapse while a session recovery is performed. If the session recovery is not over within this time, initiator terminates the session.",
                  "key": "RecoveryTimeout",
                  "optionType": {
                    "valueIsReadonly": false,
                    "min": 1,
                    "max": 120,
                    "defaultValue": 10
                  }
                },
                {
                  "label": "NoopTimeout",
                  "summary": "iSCSI option : Time in seconds that can elapse before initiator receives a NOP-IN PDU from the target. When no-op timeout limit is exceeded, initiator terminates the current connection and starts a new one.",
                  "key": "NoopTimeout",
                  "optionType": {
                    "valueIsReadonly": false,
                    "min": 10,
                    "max": 30,
                    "defaultValue": 10
                  }
                },
                {
                  "label": "NoopInterval",
                  "summary": "iSCSI option : Time in seconds in between NOP-OUTs sent by the initiator to verify that a connection is still active.",
                  "key": "NoopInterval",
                  "optionType": {
                    "valueIsReadonly": false,
                    "min": 1,
                    "max": 60,
                    "defaultValue": 15
                  }
                },
                {
                  "label": "InitR2T",
                  "summary": "iSCSI option : Whether to allow initiator to start sending data to a target as if it had received an initial R2T.",
                  "key": "InitR2T",
                  "optionType": {
                    "valueIsReadonly": true,
                    "supported": true,
                    "defaultValue": false
                  }
                },
                {
                  "label": "ImmediateData",
                  "summary": "iSCSI option : Whether to allow initiator to send immediate (unsolicited) data to the target.",
                  "key": "ImmediateData",
                  "optionType": {
                    "valueIsReadonly": true,
                    "supported": true,
                    "defaultValue": true
                  }
                },
                {
                  "label": "DelayedAck",
                  "summary": "iSCSI option : Whether to allow initiator to delay acknowledgement of received data packets.",
                  "key": "DelayedAck",
                  "optionType": {
                    "valueIsReadonly": false,
                    "supported": true,
                    "defaultValue": true
                  }
                }
              ],
              "advancedOptions": [
                {
                  "key": "ErrorRecoveryLevel",
                  "value": 0,
                  "isInherited": true
                },
                {
                  "key": "LoginRetryMax",
                  "value": 4,
                  "isInherited": true
                },
                {
                  "key": "MaxOutstandingR2T",
                  "value": 1,
                  "isInherited": true
                },
                {
                  "key": "FirstBurstLength",
                  "value": 262144,
                  "isInherited": true
                },
                {
                  "key": "MaxBurstLength",
                  "value": 262144,
                  "isInherited": true
                },
                {
                  "key": "MaxRecvDataSegLen",
                  "value": 131072,
                  "isInherited": true
                },
                {
                  "key": "MaxCommands",
                  "value": 128,
                  "isInherited": true
                },
                {
                  "key": "DefaultTimeToWait",
                  "value": 2,
                  "isInherited": true
                },
                {
                  "key": "DefaultTimeToRetain",
                  "value": 0,
                  "isInherited": true
                },
                {
                  "key": "LoginTimeout",
                  "value": 5,
                  "isInherited": true
                },
                {
                  "key": "LogoutTimeout",
                  "value": 15,
                  "isInherited": true
                },
                {
                  "key": "RecoveryTimeout",
                  "value": 10,
                  "isInherited": true
                },
                {
                  "key": "NoopTimeout",
                  "value": 10,
                  "isInherited": true
                },
                {
                  "key": "NoopInterval",
                  "value": 15,
                  "isInherited": true
                },
                {
                  "key": "InitR2T",
                  "value": false,
                  "isInherited": true
                },
                {
                  "key": "ImmediateData",
                  "value": true,
                  "isInherited": true
                },
                {
                  "key": "DelayedAck",
                  "value": true,
                  "isInherited": true
                }
              ],
              "parent": "10.136.15.233:3260"
            },
            {
              "address": "10.136.15.231",
              "port": 3260,
              "iScsiName": "iqn.2010-06.com.purestorage:flasharray.15dbd8b7f06a7c64",
              "discoveryMethod": "sendTargetMethod",
              "authenticationProperties": {
                "chapAuthEnabled": false,
                "chapAuthenticationType": "chapProhibited",
                "chapInherited": true,
                "mutualChapAuthenticationType": "chapProhibited",
                "mutualChapInherited": true
              },
              "digestProperties": {
                "headerDigestType": "digestProhibited",
                "headerDigestInherited": true,
                "dataDigestType": "digestProhibited",
                "dataDigestInherited": true
              },
              "supportedAdvancedOptions": [
                {
                  "label": "ErrorRecoveryLevel",
                  "summary": "iSCSI option : iSCSI Error Recovery Level (ERL) value that the ESX initiator would negotiate during login.",
                  "key": "ErrorRecoveryLevel",
                  "optionType": {
                    "valueIsReadonly": true,
                    "min": 0,
                    "max": 2,
                    "defaultValue": 0
                  }
                },
                {
                  "label": "LoginRetryMax",
                  "summary": "iSCSI option : Maximum number of times ESX initiator would retry login to a target, before giving up.",
                  "key": "LoginRetryMax",
                  "optionType": {
                    "valueIsReadonly": true,
                    "min": 0,
                    "max": 64,
                    "defaultValue": 4
                  }
                },
                {
                  "label": "MaxOutstandingR2T",
                  "summary": "iSCSI option : Maximum number of R2T (Ready To Transfer) PDUs, that can be outstanding for a task.",
                  "key": "MaxOutstandingR2T",
                  "optionType": {
                    "valueIsReadonly": false,
                    "min": 1,
                    "max": 8,
                    "defaultValue": 1
                  }
                },
                {
                  "label": "FirstBurstLength",
                  "summary": "iSCSI option : Maximum unsolicited data in bytes initiator can send during the execution of a single SCSI command. It must be <= MaxBurstLength.",
                  "key": "FirstBurstLength",
                  "optionType": {
                    "valueIsReadonly": false,
                    "min": 512,
                    "max": 16777215,
                    "defaultValue": 262144
                  }
                },
                {
                  "label": "MaxBurstLength",
                  "summary": "iSCSI option : Maximum SCSI data payload in bytes in a Data-In or a solicited Data-Out iSCSI sequence.",
                  "key": "MaxBurstLength",
                  "optionType": {
                    "valueIsReadonly": false,
                    "min": 512,
                    "max": 16777215,
                    "defaultValue": 262144
                  }
                },
                {
                  "label": "MaxRecvDataSegLen",
                  "summary": "iSCSI option : Maximum data segment length in bytes that can be received in an iSCSI PDU. It is recommended to keep it <= MaxBurstLength.",
                  "key": "MaxRecvDataSegLen",
                  "optionType": {
                    "valueIsReadonly": false,
                    "min": 512,
                    "max": 16777215,
                    "defaultValue": 131072
                  }
                },
                {
                  "label": "MaxCommands",
                  "summary": "iSCSI option : Maximum SCSI commands that can be queued on the iscsi adpater.",
                  "key": "MaxCommands",
                  "optionType": {
                    "valueIsReadonly": true,
                    "min": 2,
                    "max": 2048,
                    "defaultValue": 128
                  }
                },
                {
                  "label": "DefaultTimeToWait",
                  "summary": "iSCSI option : Minimum seconds to wait before attempting a logout or an active task reassignment after an unexpected connection termination or reset.",
                  "key": "DefaultTimeToWait",
                  "optionType": {
                    "valueIsReadonly": true,
                    "min": 0,
                    "max": 60,
                    "defaultValue": 2
                  }
                },
                {
                  "label": "DefaultTimeToRetain",
                  "summary": "iSCSI option : Maximum seconds that a connection and task allegiance reinstatement is still possible following a connection termination or reset.",
                  "key": "DefaultTimeToRetain",
                  "optionType": {
                    "valueIsReadonly": true,
                    "min": 0,
                    "max": 60,
                    "defaultValue": 0
                  }
                },
                {
                  "label": "LoginTimeout",
                  "summary": "iSCSI option : Time in seconds initiator will wait for the Login response to finish.",
                  "key": "LoginTimeout",
                  "optionType": {
                    "valueIsReadonly": false,
                    "min": 1,
                    "max": 60,
                    "defaultValue": 5
                  }
                },
                {
                  "label": "LogoutTimeout",
                  "summary": "iSCSI option : Time in seconds initiator will wait to get a response for Logout request PDU",
                  "key": "LogoutTimeout",
                  "optionType": {
                    "valueIsReadonly": true,
                    "min": 0,
                    "max": 60,
                    "defaultValue": 15
                  }
                },
                {
                  "label": "RecoveryTimeout",
                  "summary": "iSCSI option : Time in seconds that can elapse while a session recovery is performed. If the session recovery is not over within this time, initiator terminates the session.",
                  "key": "RecoveryTimeout",
                  "optionType": {
                    "valueIsReadonly": false,
                    "min": 1,
                    "max": 120,
                    "defaultValue": 10
                  }
                },
                {
                  "label": "NoopTimeout",
                  "summary": "iSCSI option : Time in seconds that can elapse before initiator receives a NOP-IN PDU from the target. When no-op timeout limit is exceeded, initiator terminates the current connection and starts a new one.",
                  "key": "NoopTimeout",
                  "optionType": {
                    "valueIsReadonly": false,
                    "min": 10,
                    "max": 30,
                    "defaultValue": 10
                  }
                },
                {
                  "label": "NoopInterval",
                  "summary": "iSCSI option : Time in seconds in between NOP-OUTs sent by the initiator to verify that a connection is still active.",
                  "key": "NoopInterval",
                  "optionType": {
                    "valueIsReadonly": false,
                    "min": 1,
                    "max": 60,
                    "defaultValue": 15
                  }
                },
                {
                  "label": "InitR2T",
                  "summary": "iSCSI option : Whether to allow initiator to start sending data to a target as if it had received an initial R2T.",
                  "key": "InitR2T",
                  "optionType": {
                    "valueIsReadonly": true,
                    "supported": true,
                    "defaultValue": false
                  }
                },
                {
                  "label": "ImmediateData",
                  "summary": "iSCSI option : Whether to allow initiator to send immediate (unsolicited) data to the target.",
                  "key": "ImmediateData",
                  "optionType": {
                    "valueIsReadonly": true,
                    "supported": true,
                    "defaultValue": true
                  }
                },
                {
                  "label": "DelayedAck",
                  "summary": "iSCSI option : Whether to allow initiator to delay acknowledgement of received data packets.",
                  "key": "DelayedAck",
                  "optionType": {
                    "valueIsReadonly": false,
                    "supported": true,
                    "defaultValue": true
                  }
                }
              ],
              "advancedOptions": [
                {
                  "key": "ErrorRecoveryLevel",
                  "value": 0,
                  "isInherited": true
                },
                {
                  "key": "LoginRetryMax",
                  "value": 4,
                  "isInherited": true
                },
                {
                  "key": "MaxOutstandingR2T",
                  "value": 1,
                  "isInherited": true
                },
                {
                  "key": "FirstBurstLength",
                  "value": 262144,
                  "isInherited": true
                },
                {
                  "key": "MaxBurstLength",
                  "value": 262144,
                  "isInherited": true
                },
                {
                  "key": "MaxRecvDataSegLen",
                  "value": 131072,
                  "isInherited": true
                },
                {
                  "key": "MaxCommands",
                  "value": 128,
                  "isInherited": true
                },
                {
                  "key": "DefaultTimeToWait",
                  "value": 2,
                  "isInherited": true
                },
                {
                  "key": "DefaultTimeToRetain",
                  "value": 0,
                  "isInherited": true
                },
                {
                  "key": "LoginTimeout",
                  "value": 5,
                  "isInherited": true
                },
                {
                  "key": "LogoutTimeout",
                  "value": 15,
                  "isInherited": true
                },
                {
                  "key": "RecoveryTimeout",
                  "value": 10,
                  "isInherited": true
                },
                {
                  "key": "NoopTimeout",
                  "value": 10,
                  "isInherited": true
                },
                {
                  "key": "NoopInterval",
                  "value": 15,
                  "isInherited": true
                },
                {
                  "key": "InitR2T",
                  "value": false,
                  "isInherited": true
                },
                {
                  "key": "ImmediateData",
                  "value": true,
                  "isInherited": true
                },
                {
                  "key": "DelayedAck",
                  "value": true,
                  "isInherited": true
                }
              ],
              "parent": "10.136.15.233:3260"
            },
            {
              "address": "10.136.13.231",
              "port": 3260,
              "iScsiName": "iqn.2010-06.com.purestorage:flasharray.15dbd8b7f06a7c64",
              "discoveryMethod": "sendTargetMethod",
              "authenticationProperties": {
                "chapAuthEnabled": false,
                "chapAuthenticationType": "chapProhibited",
                "chapInherited": true,
                "mutualChapAuthenticationType": "chapProhibited",
                "mutualChapInherited": true
              },
              "digestProperties": {
                "headerDigestType": "digestProhibited",
                "headerDigestInherited": true,
                "dataDigestType": "digestProhibited",
                "dataDigestInherited": true
              },
              "supportedAdvancedOptions": [
                {
                  "label": "ErrorRecoveryLevel",
                  "summary": "iSCSI option : iSCSI Error Recovery Level (ERL) value that the ESX initiator would negotiate during login.",
                  "key": "ErrorRecoveryLevel",
                  "optionType": {
                    "valueIsReadonly": true,
                    "min": 0,
                    "max": 2,
                    "defaultValue": 0
                  }
                },
                {
                  "label": "LoginRetryMax",
                  "summary": "iSCSI option : Maximum number of times ESX initiator would retry login to a target, before giving up.",
                  "key": "LoginRetryMax",
                  "optionType": {
                    "valueIsReadonly": true,
                    "min": 0,
                    "max": 64,
                    "defaultValue": 4
                  }
                },
                {
                  "label": "MaxOutstandingR2T",
                  "summary": "iSCSI option : Maximum number of R2T (Ready To Transfer) PDUs, that can be outstanding for a task.",
                  "key": "MaxOutstandingR2T",
                  "optionType": {
                    "valueIsReadonly": false,
                    "min": 1,
                    "max": 8,
                    "defaultValue": 1
                  }
                },
                {
                  "label": "FirstBurstLength",
                  "summary": "iSCSI option : Maximum unsolicited data in bytes initiator can send during the execution of a single SCSI command. It must be <= MaxBurstLength.",
                  "key": "FirstBurstLength",
                  "optionType": {
                    "valueIsReadonly": false,
                    "min": 512,
                    "max": 16777215,
                    "defaultValue": 262144
                  }
                },
                {
                  "label": "MaxBurstLength",
                  "summary": "iSCSI option : Maximum SCSI data payload in bytes in a Data-In or a solicited Data-Out iSCSI sequence.",
                  "key": "MaxBurstLength",
                  "optionType": {
                    "valueIsReadonly": false,
                    "min": 512,
                    "max": 16777215,
                    "defaultValue": 262144
                  }
                },
                {
                  "label": "MaxRecvDataSegLen",
                  "summary": "iSCSI option : Maximum data segment length in bytes that can be received in an iSCSI PDU. It is recommended to keep it <= MaxBurstLength.",
                  "key": "MaxRecvDataSegLen",
                  "optionType": {
                    "valueIsReadonly": false,
                    "min": 512,
                    "max": 16777215,
                    "defaultValue": 131072
                  }
                },
                {
                  "label": "MaxCommands",
                  "summary": "iSCSI option : Maximum SCSI commands that can be queued on the iscsi adpater.",
                  "key": "MaxCommands",
                  "optionType": {
                    "valueIsReadonly": true,
                    "min": 2,
                    "max": 2048,
                    "defaultValue": 128
                  }
                },
                {
                  "label": "DefaultTimeToWait",
                  "summary": "iSCSI option : Minimum seconds to wait before attempting a logout or an active task reassignment after an unexpected connection termination or reset.",
                  "key": "DefaultTimeToWait",
                  "optionType": {
                    "valueIsReadonly": true,
                    "min": 0,
                    "max": 60,
                    "defaultValue": 2
                  }
                },
                {
                  "label": "DefaultTimeToRetain",
                  "summary": "iSCSI option : Maximum seconds that a connection and task allegiance reinstatement is still possible following a connection termination or reset.",
                  "key": "DefaultTimeToRetain",
                  "optionType": {
                    "valueIsReadonly": true,
                    "min": 0,
                    "max": 60,
                    "defaultValue": 0
                  }
                },
                {
                  "label": "LoginTimeout",
                  "summary": "iSCSI option : Time in seconds initiator will wait for the Login response to finish.",
                  "key": "LoginTimeout",
                  "optionType": {
                    "valueIsReadonly": false,
                    "min": 1,
                    "max": 60,
                    "defaultValue": 5
                  }
                },
                {
                  "label": "LogoutTimeout",
                  "summary": "iSCSI option : Time in seconds initiator will wait to get a response for Logout request PDU",
                  "key": "LogoutTimeout",
                  "optionType": {
                    "valueIsReadonly": true,
                    "min": 0,
                    "max": 60,
                    "defaultValue": 15
                  }
                },
                {
                  "label": "RecoveryTimeout",
                  "summary": "iSCSI option : Time in seconds that can elapse while a session recovery is performed. If the session recovery is not over within this time, initiator terminates the session.",
                  "key": "RecoveryTimeout",
                  "optionType": {
                    "valueIsReadonly": false,
                    "min": 1,
                    "max": 120,
                    "defaultValue": 10
                  }
                },
                {
                  "label": "NoopTimeout",
                  "summary": "iSCSI option : Time in seconds that can elapse before initiator receives a NOP-IN PDU from the target. When no-op timeout limit is exceeded, initiator terminates the current connection and starts a new one.",
                  "key": "NoopTimeout",
                  "optionType": {
                    "valueIsReadonly": false,
                    "min": 10,
                    "max": 30,
                    "defaultValue": 10
                  }
                },
                {
                  "label": "NoopInterval",
                  "summary": "iSCSI option : Time in seconds in between NOP-OUTs sent by the initiator to verify that a connection is still active.",
                  "key": "NoopInterval",
                  "optionType": {
                    "valueIsReadonly": false,
                    "min": 1,
                    "max": 60,
                    "defaultValue": 15
                  }
                },
                {
                  "label": "InitR2T",
                  "summary": "iSCSI option : Whether to allow initiator to start sending data to a target as if it had received an initial R2T.",
                  "key": "InitR2T",
                  "optionType": {
                    "valueIsReadonly": true,
                    "supported": true,
                    "defaultValue": false
                  }
                },
                {
                  "label": "ImmediateData",
                  "summary": "iSCSI option : Whether to allow initiator to send immediate (unsolicited) data to the target.",
                  "key": "ImmediateData",
                  "optionType": {
                    "valueIsReadonly": true,
                    "supported": true,
                    "defaultValue": true
                  }
                },
                {
                  "label": "DelayedAck",
                  "summary": "iSCSI option : Whether to allow initiator to delay acknowledgement of received data packets.",
                  "key": "DelayedAck",
                  "optionType": {
                    "valueIsReadonly": false,
                    "supported": true,
                    "defaultValue": true
                  }
                }
              ],
              "advancedOptions": [
                {
                  "key": "ErrorRecoveryLevel",
                  "value": 0,
                  "isInherited": true
                },
                {
                  "key": "LoginRetryMax",
                  "value": 4,
                  "isInherited": true
                },
                {
                  "key": "MaxOutstandingR2T",
                  "value": 1,
                  "isInherited": true
                },
                {
                  "key": "FirstBurstLength",
                  "value": 262144,
                  "isInherited": true
                },
                {
                  "key": "MaxBurstLength",
                  "value": 262144,
                  "isInherited": true
                },
                {
                  "key": "MaxRecvDataSegLen",
                  "value": 131072,
                  "isInherited": true
                },
                {
                  "key": "MaxCommands",
                  "value": 128,
                  "isInherited": true
                },
                {
                  "key": "DefaultTimeToWait",
                  "value": 2,
                  "isInherited": true
                },
                {
                  "key": "DefaultTimeToRetain",
                  "value": 0,
                  "isInherited": true
                },
                {
                  "key": "LoginTimeout",
                  "value": 5,
                  "isInherited": true
                },
                {
                  "key": "LogoutTimeout",
                  "value": 15,
                  "isInherited": true
                },
                {
                  "key": "RecoveryTimeout",
                  "value": 10,
                  "isInherited": true
                },
                {
                  "key": "NoopTimeout",
                  "value": 10,
                  "isInherited": true
                },
                {
                  "key": "NoopInterval",
                  "value": 15,
                  "isInherited": true
                },
                {
                  "key": "InitR2T",
                  "value": false,
                  "isInherited": true
                },
                {
                  "key": "ImmediateData",
                  "value": true,
                  "isInherited": true
                },
                {
                  "key": "DelayedAck",
                  "value": true,
                  "isInherited": true
                }
              ],
              "parent": "10.136.15.233:3260"
            },
            {
              "address": "10.136.15.230",
              "port": 3260,
              "iScsiName": "iqn.2010-06.com.purestorage:flasharray.15dbd8b7f06a7c64",
              "discoveryMethod": "sendTargetMethod",
              "authenticationProperties": {
                "chapAuthEnabled": false,
                "chapAuthenticationType": "chapProhibited",
                "chapInherited": true,
                "mutualChapAuthenticationType": "chapProhibited",
                "mutualChapInherited": true
              },
              "digestProperties": {
                "headerDigestType": "digestProhibited",
                "headerDigestInherited": true,
                "dataDigestType": "digestProhibited",
                "dataDigestInherited": true
              },
              "supportedAdvancedOptions": [
                {
                  "label": "ErrorRecoveryLevel",
                  "summary": "iSCSI option : iSCSI Error Recovery Level (ERL) value that the ESX initiator would negotiate during login.",
                  "key": "ErrorRecoveryLevel",
                  "optionType": {
                    "valueIsReadonly": true,
                    "min": 0,
                    "max": 2,
                    "defaultValue": 0
                  }
                },
                {
                  "label": "LoginRetryMax",
                  "summary": "iSCSI option : Maximum number of times ESX initiator would retry login to a target, before giving up.",
                  "key": "LoginRetryMax",
                  "optionType": {
                    "valueIsReadonly": true,
                    "min": 0,
                    "max": 64,
                    "defaultValue": 4
                  }
                },
                {
                  "label": "MaxOutstandingR2T",
                  "summary": "iSCSI option : Maximum number of R2T (Ready To Transfer) PDUs, that can be outstanding for a task.",
                  "key": "MaxOutstandingR2T",
                  "optionType": {
                    "valueIsReadonly": false,
                    "min": 1,
                    "max": 8,
                    "defaultValue": 1
                  }
                },
                {
                  "label": "FirstBurstLength",
                  "summary": "iSCSI option : Maximum unsolicited data in bytes initiator can send during the execution of a single SCSI command. It must be <= MaxBurstLength.",
                  "key": "FirstBurstLength",
                  "optionType": {
                    "valueIsReadonly": false,
                    "min": 512,
                    "max": 16777215,
                    "defaultValue": 262144
                  }
                },
                {
                  "label": "MaxBurstLength",
                  "summary": "iSCSI option : Maximum SCSI data payload in bytes in a Data-In or a solicited Data-Out iSCSI sequence.",
                  "key": "MaxBurstLength",
                  "optionType": {
                    "valueIsReadonly": false,
                    "min": 512,
                    "max": 16777215,
                    "defaultValue": 262144
                  }
                },
                {
                  "label": "MaxRecvDataSegLen",
                  "summary": "iSCSI option : Maximum data segment length in bytes that can be received in an iSCSI PDU. It is recommended to keep it <= MaxBurstLength.",
                  "key": "MaxRecvDataSegLen",
                  "optionType": {
                    "valueIsReadonly": false,
                    "min": 512,
                    "max": 16777215,
                    "defaultValue": 131072
                  }
                },
                {
                  "label": "MaxCommands",
                  "summary": "iSCSI option : Maximum SCSI commands that can be queued on the iscsi adpater.",
                  "key": "MaxCommands",
                  "optionType": {
                    "valueIsReadonly": true,
                    "min": 2,
                    "max": 2048,
                    "defaultValue": 128
                  }
                },
                {
                  "label": "DefaultTimeToWait",
                  "summary": "iSCSI option : Minimum seconds to wait before attempting a logout or an active task reassignment after an unexpected connection termination or reset.",
                  "key": "DefaultTimeToWait",
                  "optionType": {
                    "valueIsReadonly": true,
                    "min": 0,
                    "max": 60,
                    "defaultValue": 2
                  }
                },
                {
                  "label": "DefaultTimeToRetain",
                  "summary": "iSCSI option : Maximum seconds that a connection and task allegiance reinstatement is still possible following a connection termination or reset.",
                  "key": "DefaultTimeToRetain",
                  "optionType": {
                    "valueIsReadonly": true,
                    "min": 0,
                    "max": 60,
                    "defaultValue": 0
                  }
                },
                {
                  "label": "LoginTimeout",
                  "summary": "iSCSI option : Time in seconds initiator will wait for the Login response to finish.",
                  "key": "LoginTimeout",
                  "optionType": {
                    "valueIsReadonly": false,
                    "min": 1,
                    "max": 60,
                    "defaultValue": 5
                  }
                },
                {
                  "label": "LogoutTimeout",
                  "summary": "iSCSI option : Time in seconds initiator will wait to get a response for Logout request PDU",
                  "key": "LogoutTimeout",
                  "optionType": {
                    "valueIsReadonly": true,
                    "min": 0,
                    "max": 60,
                    "defaultValue": 15
                  }
                },
                {
                  "label": "RecoveryTimeout",
                  "summary": "iSCSI option : Time in seconds that can elapse while a session recovery is performed. If the session recovery is not over within this time, initiator terminates the session.",
                  "key": "RecoveryTimeout",
                  "optionType": {
                    "valueIsReadonly": false,
                    "min": 1,
                    "max": 120,
                    "defaultValue": 10
                  }
                },
                {
                  "label": "NoopTimeout",
                  "summary": "iSCSI option : Time in seconds that can elapse before initiator receives a NOP-IN PDU from the target. When no-op timeout limit is exceeded, initiator terminates the current connection and starts a new one.",
                  "key": "NoopTimeout",
                  "optionType": {
                    "valueIsReadonly": false,
                    "min": 10,
                    "max": 30,
                    "defaultValue": 10
                  }
                },
                {
                  "label": "NoopInterval",
                  "summary": "iSCSI option : Time in seconds in between NOP-OUTs sent by the initiator to verify that a connection is still active.",
                  "key": "NoopInterval",
                  "optionType": {
                    "valueIsReadonly": false,
                    "min": 1,
                    "max": 60,
                    "defaultValue": 15
                  }
                },
                {
                  "label": "InitR2T",
                  "summary": "iSCSI option : Whether to allow initiator to start sending data to a target as if it had received an initial R2T.",
                  "key": "InitR2T",
                  "optionType": {
                    "valueIsReadonly": true,
                    "supported": true,
                    "defaultValue": false
                  }
                },
                {
                  "label": "ImmediateData",
                  "summary": "iSCSI option : Whether to allow initiator to send immediate (unsolicited) data to the target.",
                  "key": "ImmediateData",
                  "optionType": {
                    "valueIsReadonly": true,
                    "supported": true,
                    "defaultValue": true
                  }
                },
                {
                  "label": "DelayedAck",
                  "summary": "iSCSI option : Whether to allow initiator to delay acknowledgement of received data packets.",
                  "key": "DelayedAck",
                  "optionType": {
                    "valueIsReadonly": false,
                    "supported": true,
                    "defaultValue": true
                  }
                }
              ],
              "advancedOptions": [
                {
                  "key": "ErrorRecoveryLevel",
                  "value": 0,
                  "isInherited": true
                },
                {
                  "key": "LoginRetryMax",
                  "value": 4,
                  "isInherited": true
                },
                {
                  "key": "MaxOutstandingR2T",
                  "value": 1,
                  "isInherited": true
                },
                {
                  "key": "FirstBurstLength",
                  "value": 262144,
                  "isInherited": true
                },
                {
                  "key": "MaxBurstLength",
                  "value": 262144,
                  "isInherited": true
                },
                {
                  "key": "MaxRecvDataSegLen",
                  "value": 131072,
                  "isInherited": true
                },
                {
                  "key": "MaxCommands",
                  "value": 128,
                  "isInherited": true
                },
                {
                  "key": "DefaultTimeToWait",
                  "value": 2,
                  "isInherited": true
                },
                {
                  "key": "DefaultTimeToRetain",
                  "value": 0,
                  "isInherited": true
                },
                {
                  "key": "LoginTimeout",
                  "value": 5,
                  "isInherited": true
                },
                {
                  "key": "LogoutTimeout",
                  "value": 15,
                  "isInherited": true
                },
                {
                  "key": "RecoveryTimeout",
                  "value": 10,
                  "isInherited": true
                },
                {
                  "key": "NoopTimeout",
                  "value": 10,
                  "isInherited": true
                },
                {
                  "key": "NoopInterval",
                  "value": 15,
                  "isInherited": true
                },
                {
                  "key": "InitR2T",
                  "value": false,
                  "isInherited": true
                },
                {
                  "key": "ImmediateData",
                  "value": true,
                  "isInherited": true
                },
                {
                  "key": "DelayedAck",
                  "value": true,
                  "isInherited": true
                }
              ],
              "parent": "10.136.15.233:3260"
            },
            {
              "address": "10.136.15.232",
              "port": 3260,
              "iScsiName": "iqn.2010-06.com.purestorage:flasharray.15dbd8b7f06a7c64",
              "discoveryMethod": "sendTargetMethod",
              "authenticationProperties": {
                "chapAuthEnabled": false,
                "chapAuthenticationType": "chapProhibited",
                "chapInherited": true,
                "mutualChapAuthenticationType": "chapProhibited",
                "mutualChapInherited": true
              },
              "digestProperties": {
                "headerDigestType": "digestProhibited",
                "headerDigestInherited": true,
                "dataDigestType": "digestProhibited",
                "dataDigestInherited": true
              },
              "supportedAdvancedOptions": [
                {
                  "label": "ErrorRecoveryLevel",
                  "summary": "iSCSI option : iSCSI Error Recovery Level (ERL) value that the ESX initiator would negotiate during login.",
                  "key": "ErrorRecoveryLevel",
                  "optionType": {
                    "valueIsReadonly": true,
                    "min": 0,
                    "max": 2,
                    "defaultValue": 0
                  }
                },
                {
                  "label": "LoginRetryMax",
                  "summary": "iSCSI option : Maximum number of times ESX initiator would retry login to a target, before giving up.",
                  "key": "LoginRetryMax",
                  "optionType": {
                    "valueIsReadonly": true,
                    "min": 0,
                    "max": 64,
                    "defaultValue": 4
                  }
                },
                {
                  "label": "MaxOutstandingR2T",
                  "summary": "iSCSI option : Maximum number of R2T (Ready To Transfer) PDUs, that can be outstanding for a task.",
                  "key": "MaxOutstandingR2T",
                  "optionType": {
                    "valueIsReadonly": false,
                    "min": 1,
                    "max": 8,
                    "defaultValue": 1
                  }
                },
                {
                  "label": "FirstBurstLength",
                  "summary": "iSCSI option : Maximum unsolicited data in bytes initiator can send during the execution of a single SCSI command. It must be <= MaxBurstLength.",
                  "key": "FirstBurstLength",
                  "optionType": {
                    "valueIsReadonly": false,
                    "min": 512,
                    "max": 16777215,
                    "defaultValue": 262144
                  }
                },
                {
                  "label": "MaxBurstLength",
                  "summary": "iSCSI option : Maximum SCSI data payload in bytes in a Data-In or a solicited Data-Out iSCSI sequence.",
                  "key": "MaxBurstLength",
                  "optionType": {
                    "valueIsReadonly": false,
                    "min": 512,
                    "max": 16777215,
                    "defaultValue": 262144
                  }
                },
                {
                  "label": "MaxRecvDataSegLen",
                  "summary": "iSCSI option : Maximum data segment length in bytes that can be received in an iSCSI PDU. It is recommended to keep it <= MaxBurstLength.",
                  "key": "MaxRecvDataSegLen",
                  "optionType": {
                    "valueIsReadonly": false,
                    "min": 512,
                    "max": 16777215,
                    "defaultValue": 131072
                  }
                },
                {
                  "label": "MaxCommands",
                  "summary": "iSCSI option : Maximum SCSI commands that can be queued on the iscsi adpater.",
                  "key": "MaxCommands",
                  "optionType": {
                    "valueIsReadonly": true,
                    "min": 2,
                    "max": 2048,
                    "defaultValue": 128
                  }
                },
                {
                  "label": "DefaultTimeToWait",
                  "summary": "iSCSI option : Minimum seconds to wait before attempting a logout or an active task reassignment after an unexpected connection termination or reset.",
                  "key": "DefaultTimeToWait",
                  "optionType": {
                    "valueIsReadonly": true,
                    "min": 0,
                    "max": 60,
                    "defaultValue": 2
                  }
                },
                {
                  "label": "DefaultTimeToRetain",
                  "summary": "iSCSI option : Maximum seconds that a connection and task allegiance reinstatement is still possible following a connection termination or reset.",
                  "key": "DefaultTimeToRetain",
                  "optionType": {
                    "valueIsReadonly": true,
                    "min": 0,
                    "max": 60,
                    "defaultValue": 0
                  }
                },
                {
                  "label": "LoginTimeout",
                  "summary": "iSCSI option : Time in seconds initiator will wait for the Login response to finish.",
                  "key": "LoginTimeout",
                  "optionType": {
                    "valueIsReadonly": false,
                    "min": 1,
                    "max": 60,
                    "defaultValue": 5
                  }
                },
                {
                  "label": "LogoutTimeout",
                  "summary": "iSCSI option : Time in seconds initiator will wait to get a response for Logout request PDU",
                  "key": "LogoutTimeout",
                  "optionType": {
                    "valueIsReadonly": true,
                    "min": 0,
                    "max": 60,
                    "defaultValue": 15
                  }
                },
                {
                  "label": "RecoveryTimeout",
                  "summary": "iSCSI option : Time in seconds that can elapse while a session recovery is performed. If the session recovery is not over within this time, initiator terminates the session.",
                  "key": "RecoveryTimeout",
                  "optionType": {
                    "valueIsReadonly": false,
                    "min": 1,
                    "max": 120,
                    "defaultValue": 10
                  }
                },
                {
                  "label": "NoopTimeout",
                  "summary": "iSCSI option : Time in seconds that can elapse before initiator receives a NOP-IN PDU from the target. When no-op timeout limit is exceeded, initiator terminates the current connection and starts a new one.",
                  "key": "NoopTimeout",
                  "optionType": {
                    "valueIsReadonly": false,
                    "min": 10,
                    "max": 30,
                    "defaultValue": 10
                  }
                },
                {
                  "label": "NoopInterval",
                  "summary": "iSCSI option : Time in seconds in between NOP-OUTs sent by the initiator to verify that a connection is still active.",
                  "key": "NoopInterval",
                  "optionType": {
                    "valueIsReadonly": false,
                    "min": 1,
                    "max": 60,
                    "defaultValue": 15
                  }
                },
                {
                  "label": "InitR2T",
                  "summary": "iSCSI option : Whether to allow initiator to start sending data to a target as if it had received an initial R2T.",
                  "key": "InitR2T",
                  "optionType": {
                    "valueIsReadonly": true,
                    "supported": true,
                    "defaultValue": false
                  }
                },
                {
                  "label": "ImmediateData",
                  "summary": "iSCSI option : Whether to allow initiator to send immediate (unsolicited) data to the target.",
                  "key": "ImmediateData",
                  "optionType": {
                    "valueIsReadonly": true,
                    "supported": true,
                    "defaultValue": true
                  }
                },
                {
                  "label": "DelayedAck",
                  "summary": "iSCSI option : Whether to allow initiator to delay acknowledgement of received data packets.",
                  "key": "DelayedAck",
                  "optionType": {
                    "valueIsReadonly": false,
                    "supported": true,
                    "defaultValue": true
                  }
                }
              ],
              "advancedOptions": [
                {
                  "key": "ErrorRecoveryLevel",
                  "value": 0,
                  "isInherited": true
                },
                {
                  "key": "LoginRetryMax",
                  "value": 4,
                  "isInherited": true
                },
                {
                  "key": "MaxOutstandingR2T",
                  "value": 1,
                  "isInherited": true
                },
                {
                  "key": "FirstBurstLength",
                  "value": 262144,
                  "isInherited": true
                },
                {
                  "key": "MaxBurstLength",
                  "value": 262144,
                  "isInherited": true
                },
                {
                  "key": "MaxRecvDataSegLen",
                  "value": 131072,
                  "isInherited": true
                },
                {
                  "key": "MaxCommands",
                  "value": 128,
                  "isInherited": true
                },
                {
                  "key": "DefaultTimeToWait",
                  "value": 2,
                  "isInherited": true
                },
                {
                  "key": "DefaultTimeToRetain",
                  "value": 0,
                  "isInherited": true
                },
                {
                  "key": "LoginTimeout",
                  "value": 5,
                  "isInherited": true
                },
                {
                  "key": "LogoutTimeout",
                  "value": 15,
                  "isInherited": true
                },
                {
                  "key": "RecoveryTimeout",
                  "value": 10,
                  "isInherited": true
                },
                {
                  "key": "NoopTimeout",
                  "value": 10,
                  "isInherited": true
                },
                {
                  "key": "NoopInterval",
                  "value": 15,
                  "isInherited": true
                },
                {
                  "key": "InitR2T",
                  "value": false,
                  "isInherited": true
                },
                {
                  "key": "ImmediateData",
                  "value": true,
                  "isInherited": true
                },
                {
                  "key": "DelayedAck",
                  "value": true,
                  "isInherited": true
                }
              ],
              "parent": "10.136.15.233:3260"
            }
          ]"""

import json



print("send_targets contains %d elements"%len(json.loads(send_targets)))
print("static_targets contains %d elements"%len(json.loads(static_targets)))