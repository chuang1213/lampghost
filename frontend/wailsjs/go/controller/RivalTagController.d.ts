// Cynhyrchwyd y ffeil hon yn awtomatig. PEIDIWCH Â MODIWL
// This file is automatically generated. DO NOT EDIT
import {vo} from '../models';
import {result} from '../models';
import {entity} from '../models';
import {dto} from '../models';

export function AddRivalTag(arg1:vo.RivalTagVo):Promise<result.RtnMessage>;

export function DeleteRivalTagByID(arg1:number):Promise<result.RtnMessage>;

export function FindRivalTagList(arg1:vo.RivalTagVo):Promise<result.RtnDataList>;

export function GENERATE_RIVAL_TAG():Promise<entity.RivalTag>;

export function GENERATE_RIVAL_TAG_DTO():Promise<dto.RivalTagDto>;

export function QueryRivalTagPageList(arg1:vo.RivalTagVo):Promise<result.RtnPage>;

export function RevertRivalTagEnabledState(arg1:number):Promise<result.RtnMessage>;

export function SyncRivalTag(arg1:number):Promise<result.RtnMessage>;
