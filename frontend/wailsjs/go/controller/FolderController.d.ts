// Cynhyrchwyd y ffeil hon yn awtomatig. PEIDIWCH Â MODIWL
// This file is automatically generated. DO NOT EDIT
import {result} from '../models';
import {dto} from '../models';

export function AddFolder(arg1:string):Promise<result.RtnMessage>;

export function DelFolder(arg1:number):Promise<result.RtnMessage>;

export function DelFolderContent(arg1:number):Promise<result.RtnMessage>;

export function FindFolderTree():Promise<result.RtnDataList>;

export function GENERATOR_FOLDER_CONTENT_DTO():Promise<dto.FolderContentDto>;

export function GENERATOR_FOLDER_DTO():Promise<dto.FolderDto>;
